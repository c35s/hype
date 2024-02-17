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
	"syscall"
	"testing"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"github.com/cavaliergopher/cpio"
	"golang.org/x/sys/unix"
)

type GuestTest struct {
	Host  func(t *testing.T, runGuest func(vmm.Config))
	Guest func(t *testing.T)
}

// guest is set to the string "guest" using -ldflags "-X ..." when the test
// binary destined to become /init is built. The isGuest and isHost bools
// use this value to decide where the test is running.
var guest string

var (
	isGuest = guest == "guest"
	isHost  = !isGuest
)

var kernelBytes []byte
var initrdBytes []byte

func TestMain(m *testing.M) {
	if isGuest {
		m.Run()

		// If the tests are running in the VM, the exit code returned by Run is
		// discarded because the kernel really doesn't like it when PID 1 exits.
		// Instead, we reboot! The host can figure out what happened by parsing
		// console output.

		if err := unix.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
			panic(err)
		}

	}

	kb, err := os.ReadFile("../.build/linux/guest/arch/x86/boot/bzImage")
	if err != nil {
		panic(err)
	}

	kernelBytes = kb

	// build static test exe
	exe := new(bytes.Buffer)
	build := exec.Command("go", "test", "-c", "-o", "/dev/stdout",
		"-tags", "guest,netgo",
		"-ldflags", "-X github.com/c35s/hype/vmm_test.guest=guest", ".")
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

func (gt GuestTest) Run(t *testing.T) {
	t.Helper()

	switch {
	case isHost:
		t.Run("host", func(tt *testing.T) {
			gt.Host(tt, func(cfg vmm.Config) {
				runGuest(tt, t.Name()+"/guest", cfg)
			})
		})

	case isGuest:
		t.Run("guest", gt.Guest)
	}
}

func runGuest(t *testing.T, testName string, cfg vmm.Config) {
	t.Helper()

	if cfg.Loader != nil {
		t.Fatal("loader must be nil")
	}

	var console *virtio.ConsoleDevice
	for _, dev := range cfg.Devices {
		if c, ok := dev.(*virtio.ConsoleDevice); ok {
			console = c
		}
	}

	if console == nil {
		console = new(virtio.ConsoleDevice)
		cfg.Devices = append(cfg.Devices, console)
	}

	out := new(bytes.Buffer)
	outs := []io.Writer{out}
	if console.Out != nil {
		outs = append(outs, console.Out)
	}

	if testing.Verbose() {
		outs = append(outs, os.Stdout)
	}

	console.Out = io.MultiWriter(outs...)

	cfg.Loader = &linux.Loader{
		Kernel:  kernelBytes,
		Initrd:  initrdBytes,
		Cmdline: fmt.Sprintf("reboot=t console=hvc0 -- -test.v -test.run=^%s$", testName),
	}

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
}
