package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"golang.org/x/term"
)

func main() {

	var (
		memSize    = flag.Int("mem", 1024, "set the VM's memory size in MiB")
		kernelPath = flag.String("kernel", "bzImage", "load bzImage from file or URL")
		initrdPath = flag.String("initrd", "initrd.cpio.gz", "load initial ramdisk from file or URL")
		cmdline    = flag.String("cmdline", "console=hvc0 reboot=t", "set the kernel command line")
	)

	flag.Parse()

	bzImage, err := readURL(*kernelPath)
	if err != nil {
		panic(err)
	}

	initrd, err := readURL(*initrdPath)
	if err != nil {
		panic(err)
	}

	cfg := vmm.Config{
		MemSize: *memSize << 20,

		Devices: []virtio.DeviceHandler{
			&virtio.Console{
				In:  os.Stdin,
				Out: os.Stdout,
			},
		},

		Loader: &linux.Loader{
			Kernel:  bytes.NewReader(bzImage),
			Initrd:  bytes.NewReader(initrd),
			Cmdline: *cmdline,
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

func readURL(s string) (body []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("hype: read URL %s: %w", s, err)
		}
	}()

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "", "file":
		return os.ReadFile(u.Path)

	case "http", "https":
		res, err := http.Get(u.String())
		if err != nil {
			panic(err)
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("response status %d != %d", res.StatusCode, 200)
		}

		defer res.Body.Close()
		return io.ReadAll(res.Body)

	default:
		panic(u.Scheme)
	}
}
