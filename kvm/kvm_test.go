//go:build linux

package kvm_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/c35s/hype/kvm"
	"golang.org/x/sys/unix"
)

func TestGetAPIVersion(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	version, err := kvm.GetAPIVersion(sys)
	if err != nil {
		t.Fatal(err)
	}

	if version != kvm.StableAPIVersion {
		t.Fatalf("API version %d != %d", version, kvm.StableAPIVersion)
	}
}

func TestCreateVM(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()
}

func TestCheckExtension(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	if _, err := kvm.CheckExtension(sys, 0); err != nil {
		t.Fatal(err)
	}

	hlt, err := kvm.CheckExtension(sys, kvm.CapHLT)
	if err != nil {
		t.Fatal(err)
	}

	if hlt != 1 {
		t.Fatalf("hlt extension value %d != 1", hlt)
	}

	if len(kvm.AllCaps()) == 0 {
		t.Fatal("AllCaps is empty")
	}

	if s := fmt.Sprintf("%v", kvm.CapHLT); s != "KVM_CAP_HLT" {
		t.Fatalf("cap string %s != KVM_CAP_HLT", s)
	}
}

func TestCheckExtensionVM(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapCheckExtensionVM)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapCheckExtensionVM, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if _, err := kvm.CheckExtension(vm, 0); err != nil {
		t.Fatal(err)
	}
}

func TestGetVCPUMmapSize(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	sz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		t.Fatal(err)
	}

	if sz != 0x3000 {
		t.Fatalf("vcpu mmap size %#x != 0x3000", sz)
	}
}

func TestCreateVCPU(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	maxVCPUs, err := kvm.CheckExtension(sys, kvm.CapMaxVCPUs)
	if err != nil {
		t.Fatal(err)
	}

	if maxVCPUs < 1 {
		t.Fatalf("maxVCPUs %d < 1", maxVCPUs)
	}

	vcpus := make([]*kvm.VCPU, maxVCPUs)

	for i := range vcpus {
		if vcpus[i], err = kvm.CreateVCPU(vm, i); err != nil {
			t.Fatalf("create vcpu %d: %v", i, err)
		}
	}

	if _, err := kvm.CreateVCPU(vm, maxVCPUs); !errors.Is(err, unix.EINVAL) {
		t.Errorf("unexpected error exceeding max vcpus (%d): %v", maxVCPUs, err)
	}

	for i, vcpu := range vcpus {
		if err := vcpu.Close(); err != nil {
			t.Fatalf("close vcpu %d: %v", i, err)
		}
	}
}

func TestMmapVCPUState(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	mmapSz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		t.Fatal(err)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	vcpu, err := kvm.CreateVCPU(vm, 0)
	if err != nil {
		t.Fatal(err)
	}

	defer vcpu.Close()

	state, err := unix.Mmap(int(vcpu.Fd()), 0, mmapSz,
		unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

	if err != nil {
		t.Fatal(err)
	}

	defer unix.Munmap(state)

	if len(state) != mmapSz {
		t.Fatalf("mmaped size %d != %d", len(state), mmapSz)
	}
}

func TestRun(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	mmapSz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		t.Fatal(err)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	mem, err := unix.Mmap(-1, 0x0, 0x1000,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_NORESERVE)

	if err != nil {
		t.Fatal(err)
	}

	defer unix.Munmap(mem)

	region := &kvm.UserspaceMemoryRegion{
		GuestPhysAddr: 0x0,
		MemorySize:    uint64(len(mem)),
		UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
	}

	if err := kvm.SetUserMemoryRegion(vm, region); err != nil {
		t.Fatalf("unexpected error setting user memory region: %v", err)
	}

	vcpu, err := kvm.CreateVCPU(vm, 0)
	if err != nil {
		t.Fatal(err)
	}

	defer vcpu.Close()

	rawState, err := unix.Mmap(int(vcpu.Fd()), 0, mmapSz,
		unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

	if err != nil {
		t.Fatal(err)
	}

	defer unix.Munmap(rawState)

	state := (*kvm.VCPUState)(unsafe.Pointer(&rawState[0]))

	var regs kvm.Regs
	if err := kvm.GetRegs(vcpu, &regs); err != nil {
		t.Fatal(err)
	}

	var sregs kvm.Sregs
	if err := kvm.GetSregs(vcpu, &sregs); err != nil {
		t.Fatal(err)
	}

	regs.RIP = 0
	sregs.CS.Base = 0
	sregs.CS.Selector = 0

	if err := kvm.SetRegs(vcpu, &regs); err != nil {
		t.Fatal(err)
	}

	if err := kvm.SetSregs(vcpu, &sregs); err != nil {
		t.Fatal(err)
	}

	mem[0] = 0xf4 // hlt
	if err := kvm.Run(vcpu); err != nil {
		t.Fatal(err)
	}

	if state.ExitReason != kvm.ExitHLT {
		t.Fatalf("%v != %v", state.ExitReason, kvm.ExitHLT)
	}
}

func TestSetUserMemoryRegion(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	mem, err := unix.Mmap(0, 0, os.Getpagesize(), unix.PROT_READ, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		t.Fatal(err)
	}

	defer unix.Munmap(mem)

	region := &kvm.UserspaceMemoryRegion{
		MemorySize:    uint64(len(mem)),
		UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
	}

	if err := kvm.SetUserMemoryRegion(vm, region); err != nil {
		t.Fatalf("unexpected error setting user memory region: %v", err)
	}
}

func TestCreateIRQChip(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapIRQChip)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapIRQChip, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if err := kvm.CreateIRQChip(vm); err != nil {
		t.Fatal(err)
	}
}

func TestDeviceClosed(t *testing.T) {
	devFn := map[string]func(*kvm.System) error{
		"GetAPIVersion":   func(sys *kvm.System) error { _, err := kvm.GetAPIVersion(sys); return err },
		"CreateVM":        func(sys *kvm.System) error { _, err := kvm.CreateVM(sys); return err },
		"CheckExtension":  func(sys *kvm.System) error { _, err := kvm.CheckExtension(sys, 0); return err },
		"GetVCPUMmapSize": func(sys *kvm.System) error { _, err := kvm.GetVCPUMmapSize(sys); return err },
	}

	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	if err := sys.Close(); err != nil {
		t.Fatal(err)
	}

	for name, fn := range devFn {
		if err := fn(sys); !errors.Is(err, unix.EBADF) {
			t.Fatalf("%s: %v != EBADF", name, err)
		}
	}
}

func TestVMClosed(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	if err := vm.Close(); err != nil {
		t.Fatal(err)
	}

	vmFn := map[string]func(*kvm.VM) error{
		"CheckExtension":      func(vm *kvm.VM) error { _, err := kvm.CheckExtension(vm, 0); return err },
		"CreateVCPU":          func(vm *kvm.VM) error { _, err := kvm.CreateVCPU(vm, 0); return err },
		"SetUserMemoryRegion": func(vm *kvm.VM) error { return kvm.SetUserMemoryRegion(vm, nil) },
	}

	for name, fn := range vmFn {
		if err := fn(vm); !errors.Is(err, unix.EBADF) {
			t.Fatalf("%s: %v != EBADF", name, err)
		}
	}
}

func TestVCPUClosed(t *testing.T) {
	sys, err := kvm.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	vcpu, err := kvm.CreateVCPU(vm, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := vcpu.Close(); err != nil {
		t.Fatal(err)
	}

	vcpuFn := map[string]func(vcpu *kvm.VCPU) error{
		"Run": kvm.Run,
	}

	for name, fn := range vcpuFn {
		if err := fn(vcpu); !errors.Is(err, unix.EBADF) {
			t.Fatalf("%s: %v != EBADF", name, err)
		}
	}
}

func TestCapString(t *testing.T) {
	for _, c := range kvm.AllCaps() {
		unknown := fmt.Sprintf("Cap(%d)", c)
		if c.String() == unknown {
			t.Error(c)
		}
	}

	if s := kvm.Cap(9999).String(); s != "Cap(9999)" {
		t.Errorf("unexpected string for an unknown cap: %s", s)
	}
}

func TestExitString(t *testing.T) {
	for e := 0; e < 1000; e++ {
		s := kvm.Exit(e).String()
		if !strings.HasPrefix(s, "KVM_") && !strings.HasPrefix(s, "Exit(") {
			t.Errorf("malformed exit string for Exit(%d): %s", e, s)
		}
	}
}
