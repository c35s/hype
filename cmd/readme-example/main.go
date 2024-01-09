package main

import (
	"context"
	"os"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vm"
	"golang.org/x/term"
)

func main() {
	bzImage, err := os.Open(".build/linux/guest/arch/x86/boot/bzImage")
	if err != nil {
		panic(err)
	}

	initrd, err := os.Open(".build/initrd.cpio.gz")
	if err != nil {
		panic(err)
	}

	cfg := vm.Config{
		MMIO: []virtio.DeviceHandler{
			&virtio.Console{
				In:  os.Stdin,
				Out: os.Stdout,
			},
		},

		Loader: &linux.Loader{
			Kernel:  bzImage,
			Initrd:  initrd,
			Cmdline: "reboot=t console=hvc0 rdinit=/sbin/reboot -- -f",
		},
	}

	m, err := vm.New(cfg)
	if err != nil {
		panic(err)
	}

	if err := bzImage.Close(); err != nil {
		panic(err)
	}

	if err := initrd.Close(); err != nil {
		panic(err)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		old, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}

		defer term.Restore(int(os.Stdin.Fd()), old)
	}

	if err := m.Run(context.TODO()); err != nil {
		panic(err)
	}
}
