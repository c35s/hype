package linux_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/c35s/hype/os/linux"
	"github.com/google/go-cmp/cmp"
)

func TestMarshalBootParams(t *testing.T) {
	params := linux.BootParams{
		Hdr: linux.SetupHeader{
			Header: linux.SetupHeaderMagic,
		},
	}

	data, err := params.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != linux.ZeropageSize {
		t.Fatalf("boot params byte size %d != %d", len(data), linux.ZeropageSize)
	}

	zpg, err := linux.ParseBzImage(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&params, zpg); diff != "" {
		t.Fatalf("boot params differ: %s", diff)
	}
}

func TestUnmarshalBootParamsShort(t *testing.T) {
	data := make([]byte, linux.ZeropageSize-1)
	params := new(linux.BootParams)
	if err := params.UnmarshalBinary(data); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("%+v isn't ErrUnexpectedEOF", err)
	}
}

func TestParseBzImage(t *testing.T) {
	bzFile := "../.build/linux/guest/arch/x86/boot/bzImage"
	if _, err := os.Stat(bzFile); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("%s not found", bzFile)
	}

	bzImage, err := os.Open(bzFile)
	if err != nil {
		t.Fatal(err)
	}

	defer bzImage.Close()

	params, err := linux.ParseBzImage(bzImage)
	if err != nil {
		t.Fatal(err)
	}

	if params.Hdr.Xloadflags&0b1 == 0 {
		panic("kernel doesn't have a 64-bit entrypoint at 0x200")
	}

	if params.Hdr.Version != 0x20f {
		t.Errorf("boot protocol version %#x", params.Hdr.Version)
	}
}

func TestParseBadBzImage(t *testing.T) {
	t.Run("closed", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "")
		if err != nil {
			t.Fatal(err)
		}

		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = linux.ParseBzImage(f)
		if !errors.Is(err, os.ErrClosed) {
			t.Fatalf("%+v isn't ErrClosed", err)
		}
	})

	t.Run("short", func(t *testing.T) {
		_, err := linux.ParseBzImage(bytes.NewReader(make([]byte, linux.ZeropageSize-1)))
		if !errors.Is(err, io.EOF) {
			t.Fatalf("%+v isn't EOF", err)
		}
	})

	t.Run("bad magic", func(t *testing.T) {
		params := linux.BootParams{
			Hdr: linux.SetupHeader{
				Header: 0xaaaa,
			},
		}

		data, err := params.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		_, err = linux.ParseBzImage(bytes.NewReader(data))
		if !errors.Is(err, linux.ErrBzImageMagic) {
			t.Fatalf("%+v isn't ErrBzImageMagic", err)
		}
	})
}
