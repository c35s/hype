//go:build linux

package arch_test

import (
	"testing"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vmm/arch"
)

func TestArch(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	a, err := arch.New(sys)
	if err != nil {
		t.Fatal(err)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if err := a.SetupVM(vm); err != nil {
		t.Fatal(err)
	}

	mem := make([]byte, 1<<20)
	mrs, err := a.SetupMemory(mem)
	if err != nil {
		t.Fatal(err)
	}

	if len(mrs) == 0 {
		t.Fatal("no memory regions")
	}

	for _, mr := range mrs {
		if err := kvm.SetUserMemoryRegion(vm, &mr); err != nil {
			t.Fatalf("setting memory region @ slot %d: %v", mr.Slot, err)
		}
	}

	vc, err := kvm.CreateVCPU(vm, 0)
	if err != nil {
		t.Fatal(err)
	}

	defer vc.Close()

	if err := a.SetupVCPU(0, vc, nil); err != nil {
		t.Fatal(err)
	}
}
