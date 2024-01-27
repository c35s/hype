//go:build !guest

package vmm_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"github.com/cavaliergopher/cpio"
)

func TestConsole(t *testing.T) {
	out := runGuest(t)
	if !strings.Contains(out.String(), "hello from the guest") {
		t.Error("the guest didn't say hello")
	}
}

func runGuest(t *testing.T, extraMMIODevices ...virtio.DeviceHandler) (out *bytes.Buffer) {
	out = new(bytes.Buffer)

	cfg := vmm.Config{
		Devices: []virtio.DeviceHandler{
			&virtio.Console{
				Out: io.MultiWriter(os.Stdout, out),
			},
		},

		Loader: &linux.Loader{
			Kernel:  kernelBytes,
			Initrd:  initrdBytes,
			Cmdline: fmt.Sprintf("reboot=t console=hvc0 -- -test.v -test.run=^%s$", t.Name()),
		},
	}

	cfg.Devices = append(cfg.Devices, extraMMIODevices...)

	m, err := vmm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Run(context.Background()); err != nil {
		t.Error(err)
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "\r\nPASS\r\n") {
		t.FailNow()
	}

	return
}

var kernelBytes []byte
var initrdBytes []byte

func TestMain(m *testing.M) {
	kb, err := os.ReadFile("../.build/linux/guest/arch/x86/boot/bzImage")
	if err != nil {
		panic(err)
	}

	kernelBytes = kb

	// build static test exe
	exe := new(bytes.Buffer)
	build := exec.Command("go", "test", "-c", "-o", "/dev/stdout", "-tags", "guest,netgo", ".")
	build.Stdout = exe
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		panic(err)
	}

	// generate initrd
	ib := new(bytes.Buffer)
	zw := gzip.NewWriter(ib)
	cw := cpio.NewWriter(zw)

	err = cw.WriteHeader(&cpio.Header{
		Name: "init",
		Mode: 0755,
		Size: int64(exe.Len()),
	})

	if err != nil {
		panic(err)
	}

	if _, err := cw.Write(exe.Bytes()); err != nil {
		panic(err)
	}

	if err := cw.Close(); err != nil {
		panic(err)
	}

	if err := zw.Close(); err != nil {
		panic(err)
	}

	initrdBytes = ib.Bytes()

	os.Exit(m.Run())
}
