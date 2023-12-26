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

func (l *hltLoader) LoadMemory(info vm.Info, mem []byte) error {
	mem[0] = 0xf4 // hlt
	return nil
}

func (l *hltLoader) LoadVCPU(info vm.Info, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	regs.RIP = 0
	sregs.CS.Base = 0
	sregs.CS.Selector = 0
	return nil
}
