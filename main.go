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
	"strconv"
	"strings"

	"github.com/c35s/hype/os/linux"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
	"golang.org/x/term"
)

func main() {

	var (
		memSize    = flag.Int("mem", 1024, "set the VM's memory size in MiB")
		kernelPath = flag.String("kernel", "bzImage", "load bzImage from file or URL")
		initrdPath = flag.String("initrd", "", "load initial ramdisk from file or URL")
		cmdline    = flag.String("cmdline", "console=hvc0 reboot=t", "set the kernel command line")

		blkdev flagStrings
	)

	flag.Var(&blkdev, "block", "add a block device (multiple OK)")

	flag.Parse()

	bzImage, err := readURL(*kernelPath)
	if err != nil {
		panic(err)
	}

	ll := &linux.Loader{
		Kernel:  bytes.NewReader(bzImage),
		Cmdline: *cmdline,
	}

	if *initrdPath != "" {
		initrd, err := readURL(*initrdPath)
		if err != nil {
			panic(err)
		}

		ll.Initrd = bytes.NewReader(initrd)
	}

	cfg := vmm.Config{
		MemSize: *memSize << 20,

		Devices: []virtio.DeviceHandler{
			&virtio.Console{
				In:  os.Stdin,
				Out: os.Stdout,
			},
		},

		Loader: ll,
	}

	// block devices
	for _, s := range blkdev {
		s, ro := strings.CutSuffix(s, ":ro")
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}

		var stg virtio.BlockStorage

		switch u.Scheme {
		case "file", "":
			var flg int

			if !ro {
				flg = os.O_RDWR
			}

			f, err := os.OpenFile(u.Path, flg, 0)
			if err != nil {
				panic(err)
			}

			stg = &virtio.FileStorage{
				File: f,
			}

		case "http", "https":
			ro = true
			stg = &virtio.HTTPStorage{
				URL: u.String(),
			}

		case "mem":
			sz, err := strconv.ParseInt(u.Opaque, 10, 64)
			if err != nil {
				panic(err)
			}

			stg = &virtio.MemStorage{
				Bytes: make([]byte, sz),
			}

		default:
			panic("unsupported block storage scheme: " + u.Scheme)
		}

		cfg.Devices = append(cfg.Devices, &virtio.Block{
			ReadOnly: ro,
			Storage:  stg,
		})
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

// readURL reads body from a file path or URL.
// It supports file, http, and https schemes.
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

// flagStrings is a flag.Value that collects strings.
type flagStrings []string

func (*flagStrings) String() string {
	return ""
}

func (fs *flagStrings) Set(s string) error {
	*fs = append(*fs, s)
	return nil
}
