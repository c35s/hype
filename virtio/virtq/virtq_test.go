package virtq_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/c35s/hype/virtio/virtq"
)

var (
	nopMemAt  = func(addr uint64, len int) ([]byte, error) { return nil, nil }
	nopNotify = func() error { return nil }
)

var nopConfig = virtq.Config{
	MemAt:  nopMemAt,
	Notify: nopNotify,
}

func TestQ(t *testing.T) {
	t.Run("nil ring", func(t *testing.T) {
		q := virtq.New(nil, nil, nil, virtq.Config{})
		if c, err := q.Next(); c != nil || err != nil {
			t.Errorf("c=%v err=%v", c, err)
		}
	})

	t.Run("nothing available", func(t *testing.T) {
		ring := make([]virtq.Desc, 1)
		q := virtq.New(ring, nil, nil, virtq.Config{})
		if c, err := q.Next(); c != nil || err != nil {
			t.Errorf("c=%v err=%v", c, err)
		}
	})

	t.Run("one available", func(t *testing.T) {
		ring := []virtq.Desc{{Flags: virtq.DescFAvail | virtq.DescFWrite}}
		q := virtq.New(ring, new(virtq.EventSuppress), nil, nopConfig)

		c, err := q.Next()
		if err != nil {
			t.Fatal(err)
		}

		if c.Desc[0].Addr != ring[0].Addr {
			t.Error("chain[0] != ring[0]")
		}

		if c.Desc[0].IsRO() {
			t.Error("chain[0] is read-only")
		}

		if !c.Desc[0].IsWO() {
			t.Error("chain[0] is not write-only")
		}

		if err := c.Release(1); err != nil {
			t.Fatal(err)
		}

		if ring[0].Flags&virtq.DescFWrite == 0 {
			t.Error("DescFWrite flag is not set")
		}

		if c, err := q.Next(); c != nil || err != nil {
			t.Errorf("c=%v err=%v", c, err)
		}
	})

	t.Run("chained", func(t *testing.T) {
		ring := []virtq.Desc{
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail},
		}

		q := virtq.New(ring, new(virtq.EventSuppress), nil, nopConfig)

		c, err := q.Next()
		if err != nil {
			t.Fatal(err)
		}

		if c.Desc[0].Addr != ring[0].Addr {
			t.Error("chain[0] != ring[0]")
		}

		if len(c.Desc) != 3 {
			t.Errorf("len(chain) %d != 3", len(c.Desc))
		}

		if err := c.Release(0); err != nil {
			t.Fatal(err)
		}

		if c, err := q.Next(); c != nil || err != nil {
			t.Errorf("c=%v err=%v", c, err)
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

		q := virtq.New(ring, nil, nil, virtq.Config{
			MemAt: func(addr uint64, len int) ([]byte, error) {
				if addr != 0x1 {
					t.Errorf("descriptor addr %#x != %#x", addr, 0x1)
				}

				return buf.Bytes(), nil
			},
		})

		c, err := q.Next()
		if err != nil {
			t.Fatal(err)
		}

		if len(c.Desc) != 2 {
			t.Errorf("len(chain) %d != 2", len(c.Desc))
		}
	})

	t.Run("data", func(t *testing.T) {
		data := []byte("hello")
		ring := []virtq.Desc{{Addr: 0x1, Len: uint32(len(data)), Flags: virtq.DescFAvail}}

		q := virtq.New(ring, nil, nil, virtq.Config{
			MemAt: func(addr uint64, len int) ([]byte, error) {
				if addr != 0x1 {
					t.Errorf("descriptor addr %#x != %#x", addr, 0x1)
				}

				return data, nil
			},
		})

		c, err := q.Next()
		if err != nil {
			t.Fatal(err)
		}

		out, err := c.Buf(0)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(out, data) {
			t.Errorf("%q != %q", out, data)
		}
	})

	t.Run("data for a bad descriptor", func(t *testing.T) {
		q := virtq.New([]virtq.Desc{{Flags: virtq.DescFAvail}}, nil, nil, virtq.Config{})
		c, err := q.Next()
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("no panic")
			}
		}()

		c.Buf(-1)
		t.Fatal("unreachable")
	})
}
