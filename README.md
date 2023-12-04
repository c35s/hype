Hype is a collection of Go packages related to the Linux Kernel Virtual Machine (KVM). The long-term goal is to learn more about Linux internals, KVM, and virtio. The short-term goal is to boot a Linux guest with a virtio console on an amd64 Linux host.

- Package `kvm` provides wrappers for some KVM ioctls (without cgo)
- Package `vm` provides helpers for configuring and running a VM

## An example

Nothing works yet except the easy part (basic KVM ioctls). But here's how to create a VM, run, and immediately halt:

```go
package main

import (
	"context"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vm"
)

func main() {
	sys, err := kvm.Open()
	if err != nil {
		panic(err)
	}

	defer sys.Close()

	m, err := vm.New(vm.Config{
		Loader: &hltLoader{},
	})

	if err != nil {
		panic(err)
	}

	defer m.Close()

    // if err==nil, the VM halted
	if err := m.Run(context.Background()); err != nil {
		panic(err)
	}
}

// hltLoader loads a single HLT instruction at 0x0 and sets the instruction pointer to 0.
// The code segment base and sel are set to 0 as well, so the VM should halt immediately.
type hltLoader struct{}

func (l *hltLoader) LoadMemory(mem []byte) error {
	mem[0] = 0xf4 // hlt
	return nil
}

func (l *hltLoader) LoadVCPU(slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	regs.RIP = 0
	sregs.CS.Base = 0
	sregs.CS.Selector = 0
	return nil
}
```

## Development

### Panics

To remove a panic, write a test that causes it, then change the code to return an annotated error, write to the log, or otherwise handle the condition instead of panicking. Simply returning the original error usually isn't useful.
