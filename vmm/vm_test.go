//go:build linux

package vmm_test

import (
	"errors"
	"os"
	"testing"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vmm"
)

func TestValidateMemSize(t *testing.T) {
	badSizes := []int{
		os.Getpagesize() - 1,
		os.Getpagesize() + 1,
		vmm.MemSizeMin - os.Getpagesize(),
		vmm.MemSizeMax + os.Getpagesize(),
	}

	for _, sz := range badSizes {
		_, err := vmm.New(vmm.Config{
			Loader:  &nopLoader{},
			MemSize: sz,
		})

		if !errors.Is(err, vmm.ErrConfig) {
			t.Errorf("MemSize %d: error isn't ErrConfig: %v", sz, err)
		}
	}
}

func TestValidateMissingLoader(t *testing.T) {
	_, err := vmm.New(vmm.Config{})

	if !errors.Is(err, vmm.ErrConfig) {
		t.Errorf("error isn't ErrConfig: %v", err)
	}
}

func TestSetupVMError(t *testing.T) {
	boom := errors.New("boom")
	m, err := vmm.New(vmm.Config{
		Loader: nopLoader{},
		Arch: nopArch{
			SetupVMError: boom,
		},
	})

	if m != nil {
		t.Fatalf("vm is present: %v", m)
	}

	if !errors.Is(err, vmm.ErrSetup) {
		t.Errorf("error isn't ErrSetup: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Errorf("no boom: %v", err)
	}
}

func TestSetupMemoryError(t *testing.T) {
	boom := errors.New("boom")
	m, err := vmm.New(vmm.Config{
		Loader: nopLoader{},
		Arch: nopArch{
			SetupMemoryError: boom,
		},
	})

	if m != nil {
		t.Fatalf("vm is present: %v", m)
	}

	if !errors.Is(err, vmm.ErrSetupMemory) {
		t.Errorf("error isn't ErrSetupMemory: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Errorf("no boom: %v", err)
	}
}

func TestSetupVCPUError(t *testing.T) {
	boom := errors.New("boom")
	m, err := vmm.New(vmm.Config{
		Loader: nopLoader{},
		Arch: nopArch{
			SetupVCPUError: boom,
		},
	})

	if m != nil {
		t.Fatalf("vm is present: %v", m)
	}

	if !errors.Is(err, vmm.ErrSetupVCPU) {
		t.Errorf("error isn't ErrSetupVCPU: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Errorf("no boom: %v", err)
	}
}

func TestLoadMemoryError(t *testing.T) {
	boom := errors.New("boom")
	_, err := vmm.New(vmm.Config{
		Loader: &nopLoader{
			LoadMemoryError: boom,
		},
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, vmm.ErrLoadMemory) {
		t.Errorf("error isn't ErrLoadMemory: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Error("no boom")
	}
}

func TestLoadVCPUError(t *testing.T) {
	boom := errors.New("boom")
	_, err := vmm.New(vmm.Config{
		Loader: &nopLoader{
			LoadVCPUError: boom,
		},
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, vmm.ErrLoadVCPU) {
		t.Errorf("error isn't ErrLoadVCPU: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Error("no boom")
	}
}

type nopLoader struct {
	LoadMemoryError error
	LoadVCPUError   error
}

func (l nopLoader) LoadMemory(info vmm.VMInfo, mem []byte) error {
	return l.LoadMemoryError
}

func (l nopLoader) LoadVCPU(info vmm.VMInfo, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	return l.LoadVCPUError
}

type nopArch struct {
	SetupVMError     error
	SetupMemoryError error
	SetupVCPUError   error
}

func (a nopArch) SetupVM(vm *kvm.VM) error {
	return a.SetupVMError
}

func (a nopArch) SetupMemory(mem []byte) ([]kvm.UserspaceMemoryRegion, error) {
	return nil, a.SetupMemoryError
}

func (a nopArch) SetupVCPU(slot int, vcpu *kvm.VCPU, state *kvm.VCPUState) error {
	return a.SetupVCPUError
}
