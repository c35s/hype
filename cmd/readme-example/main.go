package main

import (
	"context"
	"os"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"golang.org/x/term"
)

func main() {
	bzImage, err := os.ReadFile(".build/linux/guest/arch/x86/boot/bzImage")
	if err != nil {
		panic(err)
	}

	initrd, err := os.ReadFile(".build/initrd.cpio.gz")
	if err != nil {
		panic(err)
	}

	cfg := vmm.Config{
		Devices: []virtio.DeviceConfig{
			&virtio.ConsoleDevice{
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

	m, err := vmm.New(cfg)
	if err != nil {
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
