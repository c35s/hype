// Package virtq partially implements packed virtqueues as described by the Virtual I/O
// Device (VIRTIO) Version 1.2 spec. Split virtqueues are not supported.
package virtq

import (
	"errors"
	"unsafe"
)

type Config struct {
	MemAt  func(addr uint64, len int) ([]byte, error)
	Notify func() error
}

// Queue is a packed virtqueue.
type Queue struct {
	cfg  Config
	ring []Desc
	drvE *EventSuppress
	devE *EventSuppress

	aidx uint16
	uidx uint16
	wrap bool
}

// Chain is a descriptor chain in a packed virtqueue.
type Chain struct {
	q    *Queue
	id   uint16
	skip uint16
	Desc []Desc
}

// Desc is a packed virtqueue descriptor.
type Desc struct {
	Addr  uint64
	Len   uint32
	ID    uint16
	Flags uint16
}

// EventSuppress is the driver or device event suppression area for a packed virtqueue.
type EventSuppress struct {
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
	eventFlagsEnable  = 0x0 // enable events
	eventFlagsDisable = 0x1 // disable events
	eventFlagsDesc    = 0x2 // enable events for a specific descriptor
	_                 = 0x3 // reserved
)

// New returns a new virtqueue backed by the given ring and event suppression areas.
func New(ring []Desc, drvE, devE *EventSuppress, cfg Config) *Queue {
	return &Queue{cfg: cfg, ring: ring, drvE: drvE, devE: devE, wrap: true}
}

// Next returns the next available descriptor chain, or nil if no descriptors
// are available. It returns an error if the queue's MemAt callback fails while
// resolving an indirect descriptor, or if an indirect descriptor has a
// malformed buffer.
func (q *Queue) Next() (avail *Chain, err error) {
	if len(q.ring) == 0 {
		return
	}

	i, ok := q.advance()

	if !ok {
		return
	}

	head := i

	c := &Chain{
		q:    q,
		id:   q.ring[head].ID,
		skip: 1,
		Desc: q.ring[head : head+1],
	}

	switch {
	case q.ring[i].Continues():
		for {
			i, ok = q.advance()

			if !ok {
				return nil, errors.New("descriptor continues but no next descriptor is available")
			}

			if !q.ring[i].Continues() {
				break
			}
		}

		c.skip = i - head + 1
		c.Desc = q.ring[head : i+1]

	case q.ring[i].IsIndirect():
		data, err := q.getBuf(q.ring[i])
		if err != nil {
			return nil, err
		}

		if len(data)%16 != 0 {
			return nil, errors.New("malformed indirect buffer")
		}

		c.Desc = unsafe.Slice((*Desc)(unsafe.Pointer(&data[0])), len(data)/16)
	}

	return c, nil
}

func (q *Queue) getBuf(d Desc) (buf []byte, err error) {
	if d.Len == 0 {
		return
	}

	buf, err = q.cfg.MemAt(d.Addr, int(d.Len))
	if err != nil {
		return
	}

	if len(buf) != int(d.Len) {
		return nil, errors.New("short buffer")
	}

	return
}

func (q *Queue) advance() (index uint16, ok bool) {
	a := q.ring[q.aidx].Flags&DescFAvail != 0
	u := q.ring[q.aidx].Flags&DescFUsed != 0
	if a == u || a != q.wrap {
		return
	}

	index = q.aidx
	ok = true

	q.aidx++
	if q.aidx == uint16(len(q.ring)) {
		q.aidx = 0
	}

	return
}

func (q *Queue) release(c *Chain, bytesWritten int) error {
	d := &q.ring[q.uidx]
	a := d.Flags&DescFAvail != 0
	u := d.Flags&DescFUsed != 0
	if a == u || a != q.wrap {
		panic("ring full")
	}

	var flags uint16

	if q.wrap {
		flags |= 1<<7 | 1<<15
	}

	if bytesWritten > 0 {
		flags |= DescFWrite
	}

	*d = Desc{
		ID:    c.id,
		Len:   uint32(bytesWritten),
		Flags: flags,
	}

	uidx := q.uidx
	wrap := q.wrap

	q.uidx += c.skip
	if q.uidx >= uint16(len(q.ring)) {
		q.uidx -= uint16(len(q.ring))
		q.wrap = !q.wrap
	}

	if q.drvE.ShouldNotify(uidx, wrap) {
		return q.cfg.Notify()
	}

	return nil
}

// Buf returns a slice aliasing the buffer described by the descriptor at the
// given index. It panics if the index is out of range. If the queue's MemAt
// callback fails, Buf returns the error.
func (c *Chain) Buf(i int) ([]byte, error) {
	return c.q.getBuf(c.Desc[i])
}

// Release marks the chain as used, recording the number of bytes written to the
// chain. It returns an error if the queue's Notify callback fails.
func (c *Chain) Release(bytesWritten int) error {
	return c.q.release(c, bytesWritten)
}

// Continues returns true if the descriptor's buffer continues in the next descriptor.
func (d Desc) Continues() bool {
	return d.Flags&DescFNext != 0
}

// IsWO returns true if the descriptor's buffer is device write-only.
func (d Desc) IsWO() bool {
	return d.Flags&DescFWrite != 0
}

// IsRO returns true if the descriptor's buffer is device read-only.
func (d Desc) IsRO() bool {
	return !d.IsWO()
}

// IsIndirect returns true if the descriptor's buffer contains more descriptors.
func (d Desc) IsIndirect() bool {
	return d.Flags&DescFIndirect != 0
}

// ShouldNotify returns true if a notification should be sent for the given
// descriptor index and wrap counter, or false if the event is suppressed.
func (e EventSuppress) ShouldNotify(index uint16, wrap bool) bool {
	return e.Flags == eventFlagsEnable || (e.Flags == eventFlagsDesc &&
		e.Desc&^(1<<15) == index && (e.Desc>>15 == 1) == wrap)
}
