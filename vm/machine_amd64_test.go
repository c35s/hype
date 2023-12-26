//go:build linux && amd64

package vm_test

import (
	"context"
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

type hltLoader struct{}

func (l *hltLoader) LoadMemory(_ vm.Info, mem []byte) error {
	mem[0] = 0xf4 // hlt
	return nil
}

func (l *hltLoader) LoadVCPU(_ vm.Info, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	regs.RIP = 0
	sregs.CS.Base = 0
	sregs.CS.Selector = 0
	return nil
}
