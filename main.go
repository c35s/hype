package main

import (
	"context"
	"flag"
	"os"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vm"
	"golang.org/x/term"
)

func main() {

	var (
		kernelPath = flag.String("kernel", "bzImage", "load bzImage from file")
		initrdPath = flag.String("initrd", "initrd.cpio.gz", "use archive as initial ramdisk")
		cmdline    = flag.String("cmdline", "console=hvc0 reboot=t", "set the kernel command line")
	)

	flag.Parse()

	bzImage, err := os.Open(*kernelPath)
	if err != nil {
		panic(err)
	}

	initrd, err := os.Open(*initrdPath)
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
			Cmdline: *cmdline,
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
