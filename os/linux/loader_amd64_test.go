//go:build linux && amd64

package linux_test

import (
	"context"
	"os"
	"testing"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/vm"
)

func TestLoaderReboot(t *testing.T) {
	bzImage, err := os.Open("../../.build/linux/guest/arch/x86/boot/bzImage")
	if err != nil {
		t.Fatal(err)
	}

	initrd, err := os.Open("../../.build/initrd.cpio.gz")
	if err != nil {
		t.Fatal(err)
	}

	cfg := vm.Config{
		Loader: &linux.Loader{
			Kernel:  bzImage,
			Initrd:  initrd,
			Cmdline: "reboot=t rdinit=/sbin/reboot -- -f",
		},
	}

	m, err := vm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := bzImage.Close(); err != nil {
		t.Fatal(err)
	}

	if err := initrd.Close(); err != nil {
		t.Fatal(err)
	}

	if err := m.Run(context.Background()); err != nil {
		t.Error(err)
	}
}
