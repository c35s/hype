//go:build linux

package arch

import (
	"unsafe"

	"github.com/c35s/hype/kvm"
)

type Arch struct {
	supportedCPUID []kvm.CPUIDEntry2
}

const (
	MMIOHoleAddr      = 0x0d0000000
	AfterMMIOHoleAddr = 0x100000000
)

func New(sys *kvm.System) (*Arch, error) {
	supp, err := kvm.GetSupportedCPUID(sys)
	if err != nil {
		return nil, err
	}

	a := Arch{
		supportedCPUID: supp,
	}

	return &a, nil
}

func (*Arch) SetupVM(vm *kvm.VM) error {
	return nil
}

// SetupMemory partitions mem into regions. If mem is larger than 3G, it is
// split into two regions with a 1G hole for mmio devies at MMIOHoleAddr.
func (*Arch) SetupMemory(mem []byte) ([]kvm.UserspaceMemoryRegion, error) {
	rr := []kvm.UserspaceMemoryRegion{
		{
			Slot:          0,
			GuestPhysAddr: 0,
			MemorySize:    uint64(cap(mem)),
			UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
		},
	}

	if cap(mem) > MMIOHoleAddr {
		rr = []kvm.UserspaceMemoryRegion{
			{
				Slot:          0,
				GuestPhysAddr: 0,
				MemorySize:    MMIOHoleAddr,
				UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
			},
			{
				Slot:          1,
				GuestPhysAddr: AfterMMIOHoleAddr,
				MemorySize:    uint64(cap(mem) - MMIOHoleAddr),
				UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[MMIOHoleAddr]))),
			},
		}
	}

	return rr, nil
}

// SetupVCPU sets the VCPU's cpuid to the default cpuid supported by KVM.
func (a *Arch) SetupVCPU(slot int, vcpu *kvm.VCPU, state *kvm.VCPUState) error {
	if err := kvm.SetCPUID2(vcpu, a.supportedCPUID); err != nil {
		return err
	}

	return nil
}
