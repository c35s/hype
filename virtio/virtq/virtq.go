// Package virtq partially implements packed virtqueues as described by the Virtual I/O
// Device (VIRTIO) Version 1.2 spec. Split virtqueues are not supported.
package virtq

import "unsafe"

// Queue is a packed virtqueue.
type Queue struct {
	ring []Desc
	drvE *EventSuppress
	devE *EventSuppress

	memAt  func(addr uint64, len uint32) []byte
	notify func()

	avl  uint16
	used uint16
	wrap bool
}

// Chain is a descriptor chain in a packed virtqueue.
type Chain struct {
	q    *Queue
	id   uint16
	skip uint16
	desc []Desc
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

// New returns a new packed virtqueue backed by the given descriptor ring and event
// suppression areas. The given memory access and notification callbacks are called to
// resolve descriptor data and send driver notifications.
func New(ring []Desc, drvE, devE *EventSuppress,
	memAt func(addr uint64, len uint32) []byte,
	notify func()) *Queue {

	return &Queue{
		ring:   ring,
		drvE:   drvE,
		devE:   devE,
		memAt:  memAt,
		notify: notify,
		wrap:   true,
	}
}

// Next returns the next available descriptor chain. If nothing is available, Next returns
// nil. The returned pointer is only valid until Next is called again. The caller must
// call the chain's Release method before calling Next again.
func (q *Queue) Next() *Chain {
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
		if data := q.memAt(q.ring[i].Addr, q.ring[i].Len); len(data) > 0 && len(data)%16 == 0 {
			desc = unsafe.Slice((*Desc)(unsafe.Pointer(&data[0])), len(data)/16)
		}
	}

	return &Chain{
		q:    q,
		id:   q.ring[i].ID,
		skip: uint16(skip),
		desc: desc,
	}
}

func (q *Queue) advance() (index uint16, ok bool) {
	a := q.ring[q.avl].Flags&DescFAvail != 0
	u := q.ring[q.avl].Flags&DescFUsed != 0
	if a == u || a != q.wrap {
		return
	}

	index = q.avl
	ok = true

	q.avl++
	if q.avl == uint16(len(q.ring)) {
		q.avl = 0
	}

	return
}

func (q *Queue) release(c *Chain, bytesWritten int) {
	d := &q.ring[q.used]
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

	uidx := q.used
	wrap := q.wrap

	q.used += c.skip
	if q.used >= uint16(len(q.ring)) {
		q.used -= uint16(len(q.ring))
		q.wrap = !q.wrap
	}

	if !q.drvE.isSuppressed(uidx, wrap) {
		q.notify()
	}
}

// Len returns the number of descriptors in the chain.
func (c *Chain) Len() int {
	return len(c.desc)
}

// Desc returns the indexed descriptor.
// It panics if the index is out of range.
func (c *Chain) Desc(index int) Desc {
	return c.desc[index]
}

// Data returns a slice aliasing the indexed descriptor's data.
// The slice is valid until Release is called.
// Data panics if the index is out of range.
func (c *Chain) Data(index int) []byte {
	return c.q.memAt(c.desc[index].Addr, c.desc[index].Len)
}

// IsRO returns true if the indexed descriptor is device read-only.
// It panics if the index is out of range.
func (c *Chain) IsRO(index int) bool {
	return !c.IsWO(index)
}

// IsWO returns true if the indexed descriptor is device write-only.
// It panics if the index is out of range.
func (c *Chain) IsWO(index int) bool {
	return c.desc[index].Flags&DescFWrite != 0
}

// Release marks the chain as used.
// It must be called exactly once.
func (c *Chain) Release(bytesWritten int) {
	c.q.release(c, bytesWritten)
}

func (e *EventSuppress) isSuppressed(index uint16, wrap bool) bool {
	return !(e.Flags == 0 || (e.Flags == eventFlagsDesc &&
		e.Desc&^(1<<15) == index &&
		(e.Desc>>15 == 1) == wrap))
}
