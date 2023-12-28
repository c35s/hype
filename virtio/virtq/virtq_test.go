package virtq_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/c35s/hype/virtio/virtq"
)

func TestQ(t *testing.T) {
	t.Run("nil ring", func(t *testing.T) {
		q := virtq.New(nil, nil, nil, nil)
		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("nothing available", func(t *testing.T) {
		ring := make([]virtq.D, 1)
		q := virtq.New(ring, nil, nil, nil)
		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("one available", func(t *testing.T) {
		ring := []virtq.D{{Flags: virtq.DescFAvail}}
		q := virtq.New(ring, new(virtq.E), nil, nil)

		c := q.Next()
		if &c.Desc[0] != &ring[0] {
			t.Error("&chain[0] != &ring[0]")
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
		ring := []virtq.D{
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail | virtq.DescFNext},
			{Flags: virtq.DescFAvail},
		}

		q := virtq.New(ring, new(virtq.E), nil, nil)

		c := q.Next()
		if &c.Desc[0] != &ring[0] {
			t.Error("&chain[0] != &ring[0]")
		}

		if len(c.Desc) != 3 {
			t.Errorf("len(chain) %d != 3", len(c.Desc))
		}

		c.Release(0)

		if c := q.Next(); c != nil {
			t.Errorf("chain != nil: %#v", c)
		}
	})

	t.Run("indirect", func(t *testing.T) {
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, make([]virtq.D, 2)); err != nil {
			t.Fatal(err)
		}

		ring := []virtq.D{
			{Addr: 0x1, Len: uint32(buf.Len()), Flags: virtq.DescFAvail | virtq.DescFIndirect},
		}

		q := virtq.New(ring, nil, nil, func(d *virtq.D) []byte {
			if d.Addr != 0x1 {
				t.Errorf("descriptor addr %#x != %#x", d.Addr, 0x1)
			}

			return buf.Bytes()
		})

		c := q.Next()
		if len(c.Desc) != 2 {
			t.Errorf("len(chain) %d != 2", len(c.Desc))
		}
	})

	t.Run("bytes", func(t *testing.T) {
		data := []byte("hello")
		ring := []virtq.D{{Addr: 0x1, Len: uint32(len(data)), Flags: virtq.DescFAvail}}

		q := virtq.New(ring, nil, nil, func(d *virtq.D) []byte {
			if d.Addr != 0x1 {
				t.Errorf("descriptor addr %#x != %#x", d.Addr, 0x1)
			}

			return data
		})

		c := q.Next()

		out := c.Bytes(&c.Desc[0])
		if !bytes.Equal(out, data) {
			t.Errorf("%q != %q", out, data)
		}
	})

	t.Run("bytes for a bad descriptor", func(t *testing.T) {
		q := virtq.New([]virtq.D{{Flags: virtq.DescFAvail}}, nil, nil, nil)
		c := q.Next()

		if b := c.Bytes(new(virtq.D)); b != nil {
			t.Error("bytes != nil")
		}
	})
}
