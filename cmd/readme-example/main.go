package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/c35s/hype/kvm"
	"golang.org/x/sys/unix"
)

func main() {
	sys, err := kvm.Open()
	if err != nil {
		panic(err)
	}

	defer sys.Close()

	mmapSz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		panic(err)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		panic(err)
	}

	defer vm.Close()

	mem, err := unix.Mmap(-1, 0x0, os.Getpagesize(),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_NORESERVE)

	if err != nil {
		panic(err)
	}

	defer unix.Munmap(mem)

	region := &kvm.UserspaceMemoryRegion{
		GuestPhysAddr: 0x0,
		MemorySize:    uint64(len(mem)),
		UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
	}

	if err := kvm.SetUserMemoryRegion(vm, region); err != nil {
		panic(err)
	}

	vcpu, err := kvm.CreateVCPU(vm, 0)
	if err != nil {
		panic(err)
	}

	defer vcpu.Close()

	rawState, err := unix.Mmap(int(vcpu.Fd()), 0, mmapSz,
		unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

	if err != nil {
		panic(err)
	}

	defer unix.Munmap(rawState)

	state := (*kvm.VCPUState)(unsafe.Pointer(&rawState[0]))

	var regs kvm.Regs
	if err := kvm.GetRegs(vcpu, &regs); err != nil {
		panic(err)
	}

	var sregs kvm.Sregs
	if err := kvm.GetSregs(vcpu, &sregs); err != nil {
		panic(err)
	}

	regs.RIP = 0
	sregs.CS.Base = 0
	sregs.CS.Selector = 0

	if err := kvm.SetRegs(vcpu, &regs); err != nil {
		panic(err)
	}

	if err := kvm.SetSregs(vcpu, &sregs); err != nil {
		panic(err)
	}

	mem[0] = 0xf4 // hlt
	if err := kvm.Run(vcpu); err != nil {
		panic(err)
	}

	fmt.Println(state.ExitReason)
}
