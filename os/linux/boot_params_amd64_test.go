package linux_test

import (
	"errors"
	"io"
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

	var zpg linux.BootParams
	if err := zpg.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&params, &zpg); diff != "" {
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
