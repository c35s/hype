// Package virtq partially implements packed virtqueues as described by the Virtual I/O
// Device (VIRTIO) Version 1.2 spec. Split virtqueues are not supported.
package virtq

import "unsafe"

// Q is a packed virtqueue.
type Q struct {
	ring []D
	drvE *E
	devE *E

	avlIdx  uint16
	usedIdx uint16
	wrapFlg bool

	getBytes func(*D) []byte
}

// C is a chain of descriptors in a packed virtqueue.
type C struct {
	q    *Q
	id   uint16
	skip uint16
	Desc []D
}

// D is a descriptor in a packed virtqueue.
type D struct {
	Addr  uint64
	Len   uint32
	ID    uint16
	Flags uint16
}

// E is an event suppression area for a packed virtqueue.
type E struct {
	Desc  uint16
	Flags uint16
}

const (
	DescFNext     = 1 // buffer continues in the next descriptor
	DescFWrite    = 2 // buffer is device wo (otherwise ro)
	DescFIndirect = 4 // buffer contains a descriptor table
	DescFAvail    = 1 << 7
	DescFUsed     = 1 << 15
)

const (
	ringEventFlagsEnable  = 0x0 // enable events
	ringEventFlagsDisable = 0x1 // disable events
	ringEventFlagsDesc    = 0x2 // enable events for a specific descriptor
	_                     = 0x3 // reserved
)

// New returns a new packed virtqueue backed by the given descriptor ring and event suppression areas.
// The getBytes function is called when Next needs to resolve an indirect descriptor.
func New(ring []D, drvE, devE *E, getBytes func(*D) []byte) *Q {
	return &Q{
		ring:     ring,
		drvE:     drvE,
		devE:     devE,
		wrapFlg:  true,
		getBytes: getBytes,
	}
}

// Next returns the next available descriptor chain or nil if no descriptors are
// available. A chain contains at least 1 descriptor. The caller must call the returned
// chain's Release method before calling Next again. The returned chain is only valid
// until its Release method is called.
func (q *Q) Next() *C {
	if q.ring == nil {
		return nil
	}

	i, ok := q.advance()

	if !ok {
		return nil
	}

	var (
		head = i
		skip = uint16(1)
		desc = q.ring[i : i+1]
	)

	switch {

	// chain continues w the next descriptor
	case q.ring[i].Flags&DescFNext != 0:
		for {
			if i, ok = q.advance(); !ok || q.ring[i].Flags&DescFNext == 0 {
				break
			}
		}

		skip = i - head + 1
		desc = q.ring[head : i+1]

	// chained descriptors are out-of-band
	case q.ring[i].Flags&DescFIndirect != 0:
		if data := q.getBytes(&q.ring[i]); len(data) > 0 && len(data)%16 == 0 {
			desc = unsafe.Slice((*D)(unsafe.Pointer(&data[0])), len(data)/16)
		}
	}

	return &C{
		q:    q,
		id:   q.ring[i].ID,
		skip: uint16(skip),
		Desc: desc,
	}
}

func (q *Q) advance() (index uint16, ok bool) {
	a := q.ring[q.avlIdx].Flags&DescFAvail != 0
	u := q.ring[q.avlIdx].Flags&DescFUsed != 0
	if a == u || a != q.wrapFlg {
		return
	}

	index = q.avlIdx
	ok = true

	q.avlIdx++
	if q.avlIdx == uint16(len(q.ring)) {
		q.avlIdx = 0
	}

	return
}

func (q *Q) release(c *C, bytesWritten int) (notify bool) {
	d := &q.ring[q.usedIdx]
	a := d.Flags&DescFAvail != 0
	u := d.Flags&DescFUsed != 0
	if a == u || a != q.wrapFlg {
		panic("ring full")
	}

	var flags uint16

	if q.wrapFlg {
		flags |= 1<<7 | 1<<15
	}

	if bytesWritten > 0 {
		flags |= DescFWrite
	}

	*d = D{
		ID:    c.id,
		Len:   uint32(bytesWritten),
		Flags: flags,
	}

	uidx := q.usedIdx
	wrap := q.wrapFlg

	q.usedIdx += c.skip
	if q.usedIdx >= uint16(len(q.ring)) {
		q.usedIdx -= uint16(len(q.ring))
		q.wrapFlg = !q.wrapFlg
	}

	return !q.drvE.isSuppressed(uidx, wrap)
}

// Bytes returns a slice aliasing the given descriptor's data.
// It returns nil if the descriptor isn't part of the chain.
func (c *C) Bytes(d *D) []byte {
	if uintptr(unsafe.Pointer(d)) < uintptr(unsafe.Pointer(&c.Desc[0])) ||
		uintptr(unsafe.Pointer(d)) > uintptr(unsafe.Pointer(&c.Desc[len(c.Desc)-1])) {
		return nil
	}

	return c.q.getBytes(d)
}

// Release marks the chain as used. It returns true if the caller should notify the
// receiver of the event or false if the receiver has suppressed event notification.
func (c *C) Release(bytesWritten int) (notify bool) {
	return c.q.release(c, bytesWritten)
}

func (e *E) isSuppressed(index uint16, wrap bool) bool {
	return !(e.Flags == 0 || (e.Flags == ringEventFlagsDesc &&
		e.Desc&^(1<<15) == index &&
		(e.Desc>>15 == 1) == wrap))
}
