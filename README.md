Hype is a collection of Go packages related to the Linux Kernel Virtual Machine (KVM). The long-term goal is to learn more about Linux internals, KVM, and virtio. The short-term goal is to boot a Linux guest with a virtio console, block storage, and network access on an amd64 Linux host.

[![Go reference](https://pkg.go.dev/badge/github.com/c35s/hype.svg)](https://pkg.go.dev/github.com/c35s/hype)

- Package [`kvm`](https://pkg.go.dev/github.com/c35s/hype/kvm) provides wrappers for some KVM ioctls (without cgo)
- Package [`vmm`](https://pkg.go.dev/github.com/c35s/hype/vmm) provides helpers for configuring and running a VM
- Package [`os/linux`](https://pkg.go.dev/github.com/c35s/hype/os/linux) provides a VM loader that boots a 64-bit bzImage in long mode
- Package [`virtio`](https://pkg.go.dev/github.com/c35s/hype/virtio) implements parts of the virtio 1.2 spec (basic console, block only)

## Booting a VM

This example boots Linux with an Alpine-based initrd. A virtio console is connected to stdin and stdout. The kernel is configured to run `/sbin/reboot -f` instead of a normal init, which causes the VM to exit as soon as it boots. If you want to run this example yourself, follow the instructions in "Building the guest kernel and initrd" below. Then `go run ./cmd/readme-example`.

If you remove `rdinit=/sbin/reboot -- -f` from the loader cmdline, the guest will run an init shell in a tty instead of just rebooting.

```go
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
```

![an animation showing the output of the example code](doc/readme.gif)

### Block devices

Any number of block devices can be configured. Block storage is pluggable, so a device can be backed by memory, a sparse file, an HTTP URL, or any other type implementing the `virtio.BlockStorage` interface.

Here's how to configure the builtin block storage backends:

```go
f, err := os.OpenFile("blk.raw", os.O_RDWR, 0)
if err != nil {
	panic(err)
}

cfg := vmm.Config{
	Devices: []virtio.DeviceConfig{
		&virtio.BlockDevice{
			Storage: &virtio.MemStorage{
				Bytes: make([]byte, 0x1000),
			},
		},

		&virtio.BlockDevice{
			Storage: &virtio.FileStorage{
				File: f,
			},
		},

		&virtio.BlockDevice{
			Storage: &virtio.HTTPStorage{
				URL: "https://cdn.c35s.co/ubuntu-amd64.squashfs",
			},
		},

		// ...
	},
}
```

Use something like `truncate -s 1G blk.raw` to create a local sparse file.

## Reference

- https://docs.oasis-open.org/virtio/virtio/v1.2/virtio-v1.2.html
- https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git/tree/arch/x86/include/uapi/asm/bootparam.h
- https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git/tree/arch/x86/include/uapi/asm/kvm.h
- https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git/tree/include/uapi/linux/kvm.h
- https://wiki.osdev.org/Entering_Long_Mode_Directly
- https://wiki.osdev.org/Memory_Map_(x86)
- https://wiki.osdev.org/Setting_Up_Paging
- https://wiki.osdev.org/X86-64
- https://www.kernel.org/doc/Documentation/virtual/kvm/api.txt
- https://www.kernel.org/doc/Documentation/x86/boot.txt
- https://www.kernel.org/doc/Documentation/x86/zero-page.rst

## Development

You will need an amd64 Linux environment with:

- Linux kernel dev dependencies (to build the guest kernel)
- Docker (to build the debug initrd)
- Go 1.21.1 or better

### Building the guest kernel and debug initrd

Hype includes a guest Linux kernel configuration for tests and and debugging. To build it, first make sure the `lib/linux` submodule is cloned by running `git submodule update --init`. It's gonna take a minute. After the submodule is cloned, run `make -j $(nproc) guest` to build the guest kernel and the debug initrd. The kernel build config is copied from `etc/linux/guest`. To edit the config, run `make menuconfig-guest`.

The debug initrd is built from `etc/initrd/Dockerfile`. It starts with the alpine base image and adds an init shim that mounts a few helpful things like `/dev` before executing `/bin/sh` in a tty. Try it for manual debugging and/or good old-fashioned pokin' around.

```
# this will connect your terminal to a shell in the vm
go run . -kernel .build/linux/guest/arch/x86/boot/bzImage -initrd .build/initrd.cpio.gz
```

Your shell is PID 1, so the kernel will get very angry if you exit. Use `reboot -f` instead.

### Panics

To remove a panic, write a test that causes it, then change the code to return an annotated error, write to the log, or otherwise handle the condition instead of panicking. Simply returning the original error usually isn't useful.
