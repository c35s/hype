//go:build linux && amd64

package vm_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vm"
)

func TestMachine(t *testing.T) {
	m, err := vm.New(vm.Config{
		Loader: &hltLoader{},
	})

	if err != nil {
		t.Fatal(err)
	}

	if err := m.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMemSize(t *testing.T) {
	badSizes := []int{
		os.Getpagesize() - 1,
		os.Getpagesize() + 1,
		vm.MemSizeMin - os.Getpagesize(),
		vm.MemSizeMax + os.Getpagesize(),
	}

	for _, sz := range badSizes {
		_, err := vm.New(vm.Config{
			Loader:  &hltLoader{},
			MemSize: sz,
		})

		if !errors.Is(err, vm.ErrConfig) {
			t.Errorf("MemSize %d: error isn't ErrConfig: %v", sz, err)
		}
	}
}

func TestLoadMemoryError(t *testing.T) {
	boom := errors.New("boom")
	_, err := vm.New(vm.Config{
		Loader: &errLoader{
			memoryErr: boom,
		},
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, vm.ErrLoadMemory) {
		t.Errorf("error isn't ErrLoadMemory: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Error("no boom")
	}
}

func TestLoadVCPUError(t *testing.T) {
	boom := errors.New("boom")
	_, err := vm.New(vm.Config{
		Loader: &errLoader{
			vcpuErr: boom,
		},
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, vm.ErrLoadVCPU) {
		t.Errorf("error isn't ErrLoadVCPU: %v", err)
	}

	if !errors.Is(err, boom) {
		t.Error("no boom")
	}
}

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

type errLoader struct {
	memoryErr error
	vcpuErr   error
}

func (l *errLoader) LoadMemory(mem []byte) error {
	return l.memoryErr
}

func (l *errLoader) LoadVCPU(slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	return l.vcpuErr
}
