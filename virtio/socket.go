package virtio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/c35s/hype/virtio/virtq"
	"golang.org/x/sys/unix"
)

type SocketDevice struct {
	GuestCID int

	mu   sync.Mutex
	lis  map[sockAddr]*sockListener
	done bool
}

type sockHandler struct {
	cfg   *SocketDevice
	rxc   chan *virtq.Chain
	drf   map[sockID]*sockDriverFlow
	doneC chan struct{}
}

type sockListener struct {
	addr  sockAddr
	connC chan net.Conn
	doneC chan struct{}
}

// sockDriverFlow is a vsock connection initiated by the driver.
// Along with tracking state, it acts as a device-side net.Conn.
type sockDriverFlow struct {
	ID   sockID
	Type uint16

	h *sockHandler

	// closed by sockDeviceConn.Close
	// interrupts writes
	doneC chan struct{}

	wdMu sync.Mutex
	wdT  *time.Timer
	wdC  chan struct{}

	txr, txw *os.File

	bufAlloc uint32
	fwdCnt   atomic.Uint32

	creditUpdateC chan struct{}

	numBytesRx uint32

	// high 32 bits are buf_alloc, low 32 bits are fwd_cnt
	driverCredit atomic.Uint64

	driverShutdownRecv bool
	driverShutdownSend bool
}

// vsockHdr has the same fields (but not the same binary representation) as the
// packed C struct virtio_vsock_hdr. For read-only access to a packed struct,
// use vsockHdrView.
type vsockHdr struct {
	sockID

	Len      uint32
	Type     uint16
	Op       uint16
	Flags    uint32
	BufAlloc uint32
	FwdCnt   uint32
}

// vsockHdrView is a read-only view of a packed struct virtio_vsock_hdr.
type vsockHdrView []byte

// vsockConfig is the same as struct virtio_vsock_config.
type vsockConfig struct {
	GuestCID uint64
}

// sockID uniquely identifies a vsock connection.
type sockID struct {
	SrcCID  uint64
	DstCID  uint64
	SrcPort uint32
	DstPort uint32
}

// sockAddr is a vsock address.
type sockAddr struct {
	CID  uint64
	Port uint32
}

const (
	sockRxQ = 0
	sockTxQ = 1
	sockEvQ = 2
)

const (
	vsockTypeStream    = 1
	vsockTypeSeqPacket = 2
)

const (
	vsockOpInvalid       = 0
	vsockOpRequest       = 1
	vsockOpResponse      = 2
	vsockOpRst           = 3
	vsockOpShutdown      = 4
	vsockOpRW            = 5
	vsockOpCreditUpdate  = 6
	vsockOpCreditRequest = 7
)

const (
	vsockShutdownFRecv = 0
	vsockShutdownFSend = 1
)

// sizeofVsockHdr is the binary size of a packed struct virtio_vsock_hdr.
const sizeofVsockHdr = 44

var reservedGuestCIDs = []uint32{
	1, // local
	2, // host
}

func (cfg *SocketDevice) NewHandler() (DeviceHandler, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.GuestCID == 0 {
		cfg.GuestCID = 3
	}

	if cfg.GuestCID < 0 || cfg.GuestCID >= math.MaxUint32 || slices.Contains(reservedGuestCIDs, uint32(cfg.GuestCID)) {
		return nil, fmt.Errorf("invalid guest CID %d", cfg.GuestCID)
	}

	h := &sockHandler{
		cfg:   cfg,
		rxc:   make(chan *virtq.Chain),
		drf:   make(map[sockID]*sockDriverFlow),
		doneC: make(chan struct{}),
	}

	return h, nil
}

func (cfg *SocketDevice) Listen(cid uint64, port uint32) (net.Listener, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.done {
		return nil, &net.OpError{
			Op:   "listen",
			Net:  "vsock",
			Addr: sockAddr{cid, port},
			Err:  net.ErrClosed,
		}
	}

	addr := sockAddr{CID: cid, Port: port}

	if _, ok := cfg.lis[addr]; ok {
		return nil, &net.OpError{
			Op:   "listen",
			Net:  "vsock",
			Addr: addr,
			Err:  errors.New("address already in use"),
		}
	}

	l := &sockListener{
		addr:  addr,
		connC: make(chan net.Conn),
		doneC: make(chan struct{}),
	}

	if cfg.lis == nil {
		cfg.lis = make(map[sockAddr]*sockListener)
	}

	cfg.lis[addr] = l

	go func() {
		<-l.doneC
		cfg.mu.Lock()
		delete(cfg.lis, addr)
		cfg.mu.Unlock()
	}()

	return l, nil
}

func (cfg *SocketDevice) getListener(addr sockAddr) *sockListener {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.lis[addr]
}

func (h *sockHandler) GetType() DeviceID {
	return SocketDeviceID
}

func (h *sockHandler) GetFeatures() uint64 {
	return 0
}

func (h *sockHandler) Ready(negotiatedFeatures uint64) error {
	return nil
}

func (h *sockHandler) QueueReady(num int, q *virtq.Queue, notify <-chan struct{}) error {
	switch num {
	case sockRxQ:
		go func() {
			for range notify {
				if err := h.handleRxQ(q); err != nil {
					slog.Error("vsock rx", "error", err)
				}
			}
		}()

	case sockTxQ:
		go func() {
			for range notify {
				if err := h.handleTxQ(q); err != nil {
					slog.Error("vsock tx", "error", err)
				}
			}
		}()

	case sockEvQ:
		go func() {
			for range notify {
				if err := h.handleEvQ(q); err != nil {
					slog.Error("vsock ev", "error", err)
				}
			}
		}()
	}

	return nil
}

func (h *sockHandler) handleRxQ(q *virtq.Queue) error {
	for {
		c, err := q.Next()
		if err != nil {
			return err
		}

		if c == nil {
			return nil
		}

		if len(c.Desc) > 1 {
			panic("too many rx descriptors")
		}

		if c.Desc[0].IsRO() {
			panic("write-only rx chain")
		}

		if c.Desc[0].Len < sizeofVsockHdr {
			panic("no room for the rx header")
		}

		h.rxc <- c
	}
}

func (h *sockHandler) handleTxQ(q *virtq.Queue) error {
	for {
		c, err := q.Next()
		if err != nil {
			return err
		}

		if c == nil {
			return nil
		}

		hdr, data, err := parseTxPkt(c)
		if err != nil {
			return err
		}

		// bad socket type
		if hdr.Type() != vsockTypeStream {
			return h.rst(hdr)
		}

		f, flowExists := h.drf[hdr.ID()]

		// unknown flow
		if !flowExists && hdr.Op() != vsockOpRequest {
			return h.rst(hdr)
		}

		// duplicate flow
		if flowExists && hdr.Op() == vsockOpRequest {
			return h.rst(hdr)
		}

		if flowExists {
			f.updateCredit(hdr)
		}

		switch hdr.Op() {
		case vsockOpRequest:
			err = h.handleRequestOp(hdr)

		case vsockOpRst:
			err = h.handleRstOp()

		case vsockOpShutdown:
			err = h.handleShutdownOp(f, hdr)

		case vsockOpRW:
			err = h.handleRWOp(f, data)

		case vsockOpCreditUpdate:
			err = h.handleCreditUpdateOp(f)

		case vsockOpCreditRequest:
			err = h.handleCreditRequestOp()

		default:
			panic(hdr.Op())
		}

		if err := c.Release(0); err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}
}

func (h *sockHandler) handleRequestOp(hdr vsockHdrView) error {
	l := h.cfg.getListener(hdr.ID().DstAddr())
	if l == nil {
		return h.rst(hdr)
	}

	txr, txw, err := os.Pipe()
	if err != nil {
		return err
	}

	bufAlloc, err := unix.FcntlInt(uintptr(txw.Fd()), unix.F_GETPIPE_SZ, 0)
	if err != nil {
		return err
	}

	f := &sockDriverFlow{
		ID:   hdr.ID(),
		Type: hdr.Type(),

		h: h,

		doneC: make(chan struct{}),
		wdC:   make(chan struct{}),

		txr: txr,
		txw: txw,

		bufAlloc:      uint32(bufAlloc),
		creditUpdateC: make(chan struct{}),
	}

	f.updateCredit(hdr)

	h.drf[f.ID] = f

	select {
	case l.connC <- f:
		return h.rxCtrl(vsockHdr{
			sockID:   f.ID.Swap(),
			Type:     f.Type,
			Op:       vsockOpResponse,
			BufAlloc: f.bufAlloc,
			FwdCnt:   f.fwdCnt.Load(),
		})

	default:
		return h.rst(hdr)
	}
}

func (h *sockHandler) handleRstOp() error {
	panic("handleRstOp")
}

func (h *sockHandler) handleShutdownOp(f *sockDriverFlow, hdr vsockHdrView) error {
	f.driverShutdownRecv = f.driverShutdownRecv || hdr.Flags()&(1<<vsockShutdownFRecv) != 0
	f.driverShutdownSend = f.driverShutdownSend || hdr.Flags()&(1<<vsockShutdownFSend) != 0

	if f.driverShutdownRecv && f.driverShutdownSend {
		delete(h.drf, f.ID)
		return h.rst(hdr)
	}

	return nil
}

func (h *sockHandler) handleRWOp(f *sockDriverFlow, data []byte) error {
	n, err := f.txw.Write(data)
	if err != nil {
		return err
	}

	fwdCnt := f.fwdCnt.Add(uint32(n))

	if true /*driverNeedsCreditUpdate*/ {
		return h.rxCtrl(vsockHdr{
			sockID:   f.ID.Swap(),
			Type:     f.Type,
			Op:       vsockOpCreditUpdate,
			BufAlloc: f.bufAlloc,
			FwdCnt:   fwdCnt,
		})
	}

	return nil
}

func (h *sockHandler) handleCreditUpdateOp(f *sockDriverFlow) error {
	select {
	case f.creditUpdateC <- struct{}{}:
	default:
	}

	return nil
}

func (h *sockHandler) handleCreditRequestOp() error {
	panic("handleCreditRequestOp")
}

func (h *sockHandler) handleEvQ(q *virtq.Queue) error {
	return nil
}

func (h *sockHandler) ReadConfig(p []byte, off int) error {
	cfg := vsockConfig{
		GuestCID: uint64(h.cfg.GuestCID),
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, cfg); err != nil {
		return err
	}

	if off > buf.Len() {
		return fmt.Errorf("read config: offset %d out of range", off)
	}

	if n := copy(p, buf.Bytes()[off:]); n < len(p) {
		return fmt.Errorf("short config read: %d < %d", n, len(p))
	}

	return nil
}
func (h *sockHandler) Close() error {
	select {
	case <-h.doneC:
		return net.ErrClosed

	default:
		close(h.doneC)
	}

	h.cfg.mu.Lock()
	defer h.cfg.mu.Unlock()

	h.cfg.done = true

	return nil
}

// rxCtrl sends a header-only control packet to the driver.
func (h *sockHandler) rxCtrl(hdr vsockHdr) error {
	c := <-h.rxc
	b, err := c.Buf(0)
	if err != nil {
		return err
	}

	hdr.PutBinary(b)
	if err := c.Release(sizeofVsockHdr); err != nil {
		return err
	}

	return nil
}

// rst replies to the given request header with a vsockOpRst.
func (h *sockHandler) rst(req vsockHdrView) error {
	return h.rxCtrl(vsockHdr{
		sockID: req.ID().Swap(),
		Type:   req.Type(),
		Op:     vsockOpRst,
	})
}

func (l *sockListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.connC:
		return c, nil

	case <-l.doneC:
		return nil, &net.OpError{
			Op:   "accept",
			Net:  "vsock",
			Addr: l.addr,
			Err:  net.ErrClosed,
		}
	}
}

func (l *sockListener) Close() error {
	select {
	case <-l.doneC:
		return &net.OpError{
			Op:   "close",
			Net:  "vsock",
			Addr: l.addr,
			Err:  net.ErrClosed,
		}

	default:
		close(l.doneC)
	}

	return nil
}

func (l *sockListener) Addr() net.Addr {
	return l.addr
}

func (f *sockDriverFlow) Read(p []byte) (n int, err error) {
	return f.txr.Read(p)
}

func (f *sockDriverFlow) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		var c *virtq.Chain

		select {
		case <-f.h.doneC:
			return n, f.opError("write", net.ErrClosed)

		case <-f.doneC:
			return n, f.opError("write", net.ErrClosed)

		case <-f.wdC:
			return n, f.opError("write", os.ErrDeadlineExceeded)

		case c = <-f.h.rxc:
		}

		b, err := c.Buf(0)
		if err != nil {
			return n, err
		}

		hdr := vsockHdr{
			sockID:   f.ID.Swap(),
			Len:      uint32(min(len(p), len(b)-sizeofVsockHdr)),
			Type:     f.Type,
			Op:       vsockOpRW,
			BufAlloc: f.bufAlloc,
			FwdCnt:   f.fwdCnt.Load(),
		}

		for {
			bufSpc := f.driverCredit.Load()
			bufAlloc := uint32(bufSpc >> 32)
			fwdCnt := uint32(bufSpc)
			free := bufAlloc - (f.numBytesRx - fwdCnt)

			if hdr.Len < free {
				break
			}

			select {
			case <-f.h.doneC:
				return n, f.opError("write", net.ErrClosed)

			case <-f.doneC:
				return n, f.opError("write", net.ErrClosed)

			case <-f.wdC:
				return n, f.opError("write", os.ErrDeadlineExceeded)

			case <-f.creditUpdateC:
			}
		}

		hdr.PutBinary(b)
		copy(b[sizeofVsockHdr:], p)

		if err := c.Release(sizeofVsockHdr + int(hdr.Len)); err != nil {
			return n, err
		}

		f.numBytesRx += hdr.Len
		n += int(hdr.Len)
		p = p[hdr.Len:]
	}

	return
}

func (f *sockDriverFlow) Close() error {
	select {
	case <-f.doneC:
		return f.opError("close", net.ErrClosed)

	default:
		close(f.doneC)
	}

	return nil
}

func (f *sockDriverFlow) SetDeadline(t time.Time) error {
	if err := f.SetReadDeadline(t); err != nil {
		return err
	}
	return f.SetWriteDeadline(t)
}

func (f *sockDriverFlow) SetReadDeadline(t time.Time) error {
	select {
	case <-f.doneC:
		return f.opError("set", net.ErrClosed)

	default:
		return f.txr.SetReadDeadline(t)
	}
}

func (f *sockDriverFlow) SetWriteDeadline(t time.Time) error {
	select {
	case <-f.doneC:
		return f.opError("set", net.ErrClosed)

	default:
	}

	f.wdMu.Lock()
	defer f.wdMu.Unlock()

	if f.wdT != nil && !f.wdT.Stop() {
		<-f.wdC
	}

	f.wdT = nil

	var wDCClosed bool

	select {
	case <-f.wdC:
		wDCClosed = true
	default:
	}

	if t.IsZero() {
		if wDCClosed {
			f.wdC = make(chan struct{})
		}

		return nil
	}

	d := time.Until(t)
	if d <= 0 {
		if !wDCClosed {
			close(f.wdC)
		}

		return nil
	}

	if wDCClosed {
		f.wdC = make(chan struct{})
	}

	f.wdT = time.AfterFunc(d, func() {
		close(f.wdC)
	})

	return nil
}

func (f *sockDriverFlow) LocalAddr() net.Addr {
	return f.ID.SrcAddr()
}

func (f *sockDriverFlow) RemoteAddr() net.Addr {
	return f.ID.DstAddr()
}

func (f *sockDriverFlow) updateCredit(hdr vsockHdrView) {
	f.driverCredit.Store(uint64(hdr.BufAlloc())<<32 | uint64(hdr.FwdCnt()))
}

func (f *sockDriverFlow) opError(op string, err error) error {
	return &net.OpError{Op: op, Net: "vsock", Addr: f.ID.SrcAddr(), Err: err}
}

// newSockTxPkt returns a new tx packet backed by the given chain. The chain
// must have at least one read-only descriptor. It can optionally have a second
// read-only data descriptor.
func parseTxPkt(c *virtq.Chain) (hdr vsockHdrView, data []byte, err error) {
	if !c.Desc[0].IsRO() {
		panic("write-only hdr desc")
	}

	if c.Desc[0].Len < sizeofVsockHdr {
		panic("short hdr desc")
	}

	hb, err := c.Buf(0)
	if err != nil {
		return
	}

	hdr = vsockHdrView(hb)
	data = hb[sizeofVsockHdr:]

	if len(c.Desc) > 1 {
		if len(data) > 0 {
			panic("extra data descriptor but the header descriptor had data")
		}

		if !c.Desc[1].IsRO() {
			panic("write-only data desc")
		}

		if c.Desc[1].Len != hdr.Len() {
			panic("data desc len != hdr len")
		}

		data, err = c.Buf(1)
		if err != nil {
			return
		}
	}

	return
}

func (h vsockHdr) ID() sockID {
	return h.sockID
}

func (h vsockHdr) PutBinary(p []byte) {
	_ = p[:44]
	binary.LittleEndian.PutUint64(p[0:8], h.SrcCID)
	binary.LittleEndian.PutUint64(p[8:16], h.DstCID)
	binary.LittleEndian.PutUint32(p[16:20], h.SrcPort)
	binary.LittleEndian.PutUint32(p[20:24], h.DstPort)
	binary.LittleEndian.PutUint32(p[24:28], h.Len)
	binary.LittleEndian.PutUint16(p[28:30], h.Type)
	binary.LittleEndian.PutUint16(p[30:32], h.Op)
	binary.LittleEndian.PutUint32(p[32:36], h.Flags)
	binary.LittleEndian.PutUint32(p[36:40], h.BufAlloc)
	binary.LittleEndian.PutUint32(p[40:44], h.FwdCnt)
}

func (v vsockHdrView) SrcCID() uint64 {
	return binary.LittleEndian.Uint64(v[0:8])
}

func (v vsockHdrView) DstCID() uint64 {
	return binary.LittleEndian.Uint64(v[8:16])
}

func (v vsockHdrView) SrcPort() uint32 {
	return binary.LittleEndian.Uint32(v[16:20])
}

func (v vsockHdrView) DstPort() uint32 {
	return binary.LittleEndian.Uint32(v[20:24])
}

func (v vsockHdrView) Len() uint32 {
	return binary.LittleEndian.Uint32(v[24:28])
}

func (v vsockHdrView) Type() uint16 {
	return binary.LittleEndian.Uint16(v[28:30])
}

func (v vsockHdrView) Op() uint16 {
	return binary.LittleEndian.Uint16(v[30:32])
}

func (v vsockHdrView) Flags() uint32 {
	return binary.LittleEndian.Uint32(v[32:36])
}

func (v vsockHdrView) BufAlloc() uint32 {
	return binary.LittleEndian.Uint32(v[36:40])
}

func (v vsockHdrView) FwdCnt() uint32 {
	return binary.LittleEndian.Uint32(v[40:44])
}

func (v vsockHdrView) ID() sockID {
	return sockID{
		SrcCID:  v.SrcCID(),
		DstCID:  v.DstCID(),
		SrcPort: v.SrcPort(),
		DstPort: v.DstPort(),
	}
}

func (v vsockHdrView) SrcAddr() net.Addr {
	return v.ID().SrcAddr()
}

func (v vsockHdrView) DstAddr() net.Addr {
	return v.ID().DstAddr()
}

func (id sockID) SrcAddr() sockAddr {
	return sockAddr{
		CID:  id.SrcCID,
		Port: id.SrcPort,
	}
}

func (id sockID) DstAddr() sockAddr {
	return sockAddr{
		CID:  id.DstCID,
		Port: id.DstPort,
	}
}

// Swap returns a copy of the id with src and dst fields swapped.
func (id sockID) Swap() sockID {
	return sockID{
		SrcCID:  id.DstCID,
		DstCID:  id.SrcCID,
		SrcPort: id.DstPort,
		DstPort: id.SrcPort,
	}
}

func (sockAddr) Network() string {
	return "vsock"
}

func (a sockAddr) String() string {
	return fmt.Sprintf("%d:%d", a.CID, a.Port)
}
