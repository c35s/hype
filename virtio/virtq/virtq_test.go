package virtq_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/c35s/hype/virtio/virtq"
)

var (
	nopMemAt  = func(addr uint64, len uint32) []byte { return nil }
	nopNotify = func() {}
)

func TestQ(t *testing.T) {
	t.Run("nil ring", func(t *testing.T) {
		q := virtq.New(nil, nil, nil, nopMemAt, nopNotify)
		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("nothing available", func(t *testing.T) {
		ring := make([]virtq.Desc, 1)
		q := virtq.New(ring, nil, nil, nopMemAt, nopNotify)
		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("one available", func(t *testing.T) {
		ring := []virtq.Desc{{Flags: virtq.DescFAvail | virtq.DescFWrite}}
		q := virtq.New(ring, new(virtq.EventSuppress), nil, nopMemAt, nopNotify)

		c := q.Next()
		if c.Desc(0).Addr != ring[0].Addr {
			t.Error("chain[0] != ring[0]")
		}

		if c.IsRO(0) {
			t.Error("chain[0] is read-only")
		}

		if !c.IsWO(0) {
			t.Error("chain[0] is not write-only")
		}

		c.Release(1)

		if ring[0].Flags&virtq.DescFWrite == 0 {
			t.Error("DescFWrite flag is not set")
		}

		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("chained", func(t *testing.T) {
		ring := []virtq.Desc{
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail},
		}

		q := virtq.New(ring, new(virtq.EventSuppress), nil, nopMemAt, nopNotify)

		c := q.Next()
		if c.Desc(0).Addr != ring[0].Addr {
			t.Error("chain[0] != ring[0]")
		}

		if c.Len() != 3 {
			t.Errorf("len(chain) %d != 3", c.Len())
		}

		c.Release(0)

		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("indirect", func(t *testing.T) {
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, make([]virtq.Desc, 2)); err != nil {
			t.Fatal(err)
		}

		ring := []virtq.Desc{
			{Addr: 0x1, Len: uint32(buf.Len()), Flags: virtq.DescFAvail | virtq.DescFIndirect},
		}

		memAt := func(addr uint64, size uint32) []byte {
			if addr != 0x1 {
				t.Errorf("descriptor addr %#x != %#x", addr, 0x1)
			}

			return buf.Bytes()
		}

		q := virtq.New(ring, nil, nil, memAt, nopNotify)

		c := q.Next()
		if c.Len() != 2 {
			t.Errorf("len(chain) %d != 2", c.Len())
		}
	})

	t.Run("data", func(t *testing.T) {
		data := []byte("hello")
		ring := []virtq.Desc{{Addr: 0x1, Len: uint32(len(data)), Flags: virtq.DescFAvail}}

		memAt := func(addr uint64, size uint32) []byte {
			if addr != 0x1 {
				t.Errorf("descriptor addr %#x != %#x", addr, 0x1)
			}

			return data
		}

		q := virtq.New(ring, nil, nil, memAt, nopNotify)
		c := q.Next()

		out := c.Data(0)
		if !bytes.Equal(out, data) {
			t.Errorf("%q != %q", out, data)
		}
	})

	t.Run("data for a bad descriptor", func(t *testing.T) {
		q := virtq.New([]virtq.Desc{{Flags: virtq.DescFAvail}}, nil, nil, nopMemAt, nopNotify)
		c := q.Next()

		defer func() {
			if r := recover(); r == nil {
				t.Error("no panic")
			}
		}()

		c.Data(-1)
		t.Fatal("unreachable")
	})
}
