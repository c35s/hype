package virtio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

func TestSocketDeviceNewHandler(t *testing.T) {
	badGuestCIDs := []int{-1, math.MaxUint32 + 1}
	for _, cid := range reservedGuestCIDs {
		badGuestCIDs = append(badGuestCIDs, int(cid))
	}

	for _, cid := range badGuestCIDs {
		t.Run(fmt.Sprintf("bad guest cid %d", cid), func(t *testing.T) {
			if _, err := (&SocketDevice{GuestCID: cid}).NewHandler(); err == nil {
				t.Error("no error")
			}
		})
	}

	t.Run("default guest cid", func(t *testing.T) {
		h, err := (&SocketDevice{}).NewHandler()
		if err != nil {
			t.Fatal(err)
		}

		if h.(*sockHandler).cfg.GuestCID != 3 {
			t.Errorf("guest cid %d != 3", h.(*sockHandler).cfg.GuestCID)
		}
	})
}

func Test_sockHandlerReadConfig(t *testing.T) {
	h, err := (&SocketDevice{GuestCID: 5}).NewHandler()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("guest cid", func(t *testing.T) {
		buf := make([]byte, 8)
		if err := h.ReadConfig(buf, 0); err != nil {
			t.Fatal(err)
		}

		var c vsockConfig
		if err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &c); err != nil {
			t.Fatal(err)
		}

		if c.GuestCID != 5 {
			t.Errorf("guest cid %d != 5", c.GuestCID)
		}
	})

	t.Run("oob", func(t *testing.T) {
		buf := make([]byte, 8)
		if err := h.ReadConfig(buf, 16); err == nil {
			t.Error("no error")
		}
	})

	t.Run("short", func(t *testing.T) {
		buf := make([]byte, 8)
		if err := h.ReadConfig(buf, 4); err == nil {
			t.Error("no error")
		}
	})
}

func Test_vsockHdr(t *testing.T) {
	h := vsockHdr{
		sockID: sockID{
			SrcCID:  2,
			DstCID:  3,
			SrcPort: 4,
			DstPort: 5,
		},

		Len:      6,
		Type:     7,
		Op:       8,
		Flags:    9,
		BufAlloc: 10,
		FwdCnt:   11,
	}

	if h.ID() != h.sockID {
		t.Error("unexpected ID")
	}

	if src := (sockAddr{CID: 2, Port: 4}); h.SrcAddr() != src {
		t.Errorf("src addr %v != %v", h.SrcAddr(), src)
	}

	if dst := (sockAddr{CID: 3, Port: 5}); h.DstAddr() != dst {
		t.Errorf("dst addr %v != %v", h.DstAddr(), dst)
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
		t.Fatal(err)
	}

	if buf.Len() != sizeofVsockHdr {
		t.Errorf("buf len %d != %d", buf.Len(), sizeofVsockHdr)
	}

	view := vsockHdrView(buf.Bytes())
	if view.ID() != h.sockID {
		t.Error("view: unexpected ID")
	}

	if src := (sockAddr{CID: 2, Port: 4}); view.SrcAddr() != src {
		t.Errorf("view: src addr %v != %v", view.SrcAddr(), src)
	}

	if dst := (sockAddr{CID: 3, Port: 5}); view.DstAddr() != dst {
		t.Errorf("view: dst addr %v != %v", view.DstAddr(), dst)
	}

	if view.Len() != h.Len {
		t.Errorf("view: len %d != %d", view.Len(), h.Len)
	}

	if view.Type() != h.Type {
		t.Errorf("view: type %d != %d", view.Type(), h.Type)
	}

	if view.Op() != h.Op {
		t.Errorf("view: op %d != %d", view.Op(), h.Op)
	}

	if view.Flags() != h.Flags {
		t.Errorf("view: flags %d != %d", view.Flags(), h.Flags)
	}

	if view.BufAlloc() != h.BufAlloc {
		t.Errorf("view: bufalloc %d != %d", view.BufAlloc(), h.BufAlloc)
	}

	if view.FwdCnt() != h.FwdCnt {
		t.Errorf("view: fwdcnt %d != %d", view.FwdCnt(), h.FwdCnt)
	}
}

func Test_sockIDSwap(t *testing.T) {
	id := sockID{
		SrcCID:  2,
		DstCID:  3,
		SrcPort: 4,
		DstPort: 5,
	}

	swapped := id.Swap()

	if swapped.SrcCID != id.DstCID {
		t.Errorf("swapped SrcCID %d != %d", swapped.SrcCID, id.DstCID)
	}

	if swapped.DstCID != id.SrcCID {
		t.Errorf("swapped DstCID %d != %d", swapped.DstCID, id.SrcCID)
	}

	if swapped.SrcPort != id.DstPort {
		t.Errorf("swapped SrcPort %d != %d", swapped.SrcPort, id.DstPort)
	}

	if swapped.DstPort != id.SrcPort {
		t.Errorf("swapped DstPort %d != %d", swapped.DstPort, id.SrcPort)
	}
}

func Test_sockAddr(t *testing.T) {
	a := sockAddr{
		CID:  2,
		Port: 3,
	}

	if a.Network() != "vsock" {
		t.Errorf("%s != vsock", a.Network())
	}

	if a.String() != "2:3" {
		t.Errorf("addr string %q != 2:3", a.String())
	}
}
