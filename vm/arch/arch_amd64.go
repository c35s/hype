//go:build linux

package arch

import (
	"unsafe"

	"github.com/c35s/hype/kvm"
)

type Arch struct{}

func New(sys *kvm.System) (*Arch, error) {
	return new(Arch), nil
}

func (*Arch) SetupVM(vm *kvm.VM) error {
	return nil
}

func (*Arch) SetupMemory(mem []byte) ([]kvm.UserspaceMemoryRegion, error) {
	// FIX: shouldn't always be one big region
	// (x86 has the pci hole, &c)
	reg := kvm.UserspaceMemoryRegion{
		MemorySize:    uint64(len(mem)),
		UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
	}

	return []kvm.UserspaceMemoryRegion{reg}, nil
}

func (*Arch) SetupVCPU(slot int, vcpu *kvm.VCPU, state *kvm.VCPUState) error {
	return nil
}
