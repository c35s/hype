//go:build linux && amd64

package kvm_test

import (
	"errors"
	"os"
	"testing"

	"github.com/c35s/hype/kvm"
	"golang.org/x/sys/unix"
)

func TestGetMSRIndexList(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	indices, err := kvm.GetMSRIndexList(sys)
	if err != nil {
		t.Fatal(err)
	}

	if len(indices) == 0 {
		t.Fatal("no msrs")
	}
}

func TestGetMSRFeatureIndexList(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapGetMSRFeatures)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapGetMSRFeatures, ext)
	}

	indices, err := kvm.GetMSRFeatureIndexList(sys)
	if err != nil {
		t.Fatal(err)
	}

	if len(indices) == 0 {
		t.Fatal("no msrs")
	}
}

func TestRegs(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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

	defer vcpu.Close()

	var regs kvm.Regs
	if err := kvm.GetRegs(vcpu, &regs); err != nil {
		t.Fatal(err)
	}

	if regs.RFlags != 0x2 {
		t.Fatalf("RFlags %#x != 0x2", regs.RFlags)
	}

	regs.RAX = 0xc355
	if err := kvm.SetRegs(vcpu, &regs); err != nil {
		t.Fatal(err)
	}

	if err := kvm.GetRegs(vcpu, &regs); err != nil {
		t.Fatal(err)
	}

	if regs.RAX != 0xc355 {
		t.Fatalf("RAX %#x !=  0xc355 after SetRegs", regs.RAX)
	}
}

func TestSregs(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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

	defer vcpu.Close()

	var sregs kvm.Sregs
	if err := kvm.GetSregs(vcpu, &sregs); err != nil {
		t.Fatal(err)
	}

	if sregs.CS.Base != 0xffff0000 {
		t.Fatalf("CS.Base %#x != 0xffff0000", sregs.CS.Base)
	}

	sregs.CS.Base = 0x1000
	if err := kvm.SetSregs(vcpu, &sregs); err != nil {
		t.Fatal(err)
	}

	if err := kvm.GetSregs(vcpu, &sregs); err != nil {
		t.Fatal(err)
	}

	if sregs.CS.Base != 0x1000 {
		t.Fatalf("CS.Base %#x != 0x1000 after SetSregs", sregs.CS.Base)
	}
}

func TestMSRs(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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

	defer vcpu.Close()

	indices, err := kvm.GetMSRIndexList(sys)
	if err != nil {
		t.Fatal(err)
	}

	msrs, err := kvm.GetMSRs(vcpu, indices)
	if err != nil {
		t.Fatal(err)
	}

	entries := []kvm.MSREntry{
		{Index: msrs[0].Index, Data: 0xdeadbeef},
	}

	if err := kvm.SetMSRs(vcpu, entries); err != nil {
		t.Fatal(err)
	}

	msrs, err = kvm.GetMSRs(vcpu, []int{int(entries[0].Index)})
	if err != nil {
		t.Fatal(err)
	}

	if msrs[0].Data != 0xdeadbeef {
		t.Fatalf("MSR data %#x != 0xdeadbeef", msrs[0].Data)
	}
}

func TestSystemMSRs(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapGetMSRFeatures)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapGetMSRFeatures, ext)
	}

	indices, err := kvm.GetMSRFeatureIndexList(sys)
	if err != nil {
		t.Fatal(err)
	}

	msrs, err := kvm.GetMSRs(sys, indices)
	if err != nil {
		t.Fatal(err)
	}

	if len(msrs) != len(indices) {
		t.Fatalf("len(msrs) %d != len(indices) %d", len(msrs), len(indices))
	}
}

func TestFPU(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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

	defer vcpu.Close()

	var fpu kvm.FPU
	if err := kvm.GetFPU(vcpu, &fpu); err != nil {
		t.Fatal(err)
	}

	if fpu.FCW != 0x37f {
		t.Fatalf("FCW %#x != 0x37f", fpu.FCW)
	}

	fpu.FCW = 0x1
	if err := kvm.SetFPU(vcpu, &fpu); err != nil {
		t.Fatal(err)
	}

	if err := kvm.GetFPU(vcpu, &fpu); err != nil {
		t.Fatal(err)
	}

	if fpu.FCW != 0x1 {
		t.Fatalf("FCW %#x != 0x1 after SetFPU", fpu.FCW)
	}
}

func TestClock(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapAdjustClock)
	if err != nil {
		t.Fatal(err)
	}

	if ext < 1 {
		t.Skipf("%v is %d", kvm.CapAdjustClock, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	var data kvm.ClockData
	if err := kvm.GetClock(vm, &data); err != nil {
		t.Fatal(err)
	}

	if data.Clock == 0 {
		t.Fatal("Clock is zero")
	}

	old := data.Clock

	data.Clock = 0x0
	if err := kvm.SetClock(vm, &data); err != nil {
		t.Fatal(err)
	}

	if err := kvm.GetClock(vm, &data); err != nil {
		t.Fatal(err)
	}

	if data.Clock > old {
		t.Fatalf("Clock %#x > %#x after SetClock(0)", data.Clock, old)
	}
}

func TestSetTSSAddr(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapSetTSSAddr)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapSetTSSAddr, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if err := kvm.SetTSSAddr(vm, 0xffff); err != nil {
		t.Fatal(err)
	}
}

func TestSetIdentityMapAddr(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapSetIdentityMapAddr)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapSetIdentityMapAddr, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if err := kvm.SetIdentityMapAddr(vm, 0); err != nil {
		t.Fatal(err)
	}
}

func TestCreatePIT2(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapPIT2)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapPIT2, ext)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		t.Fatal(err)
	}

	defer vm.Close()

	if err := kvm.CreateIRQChip(vm); err != nil {
		t.Fatal(err)
	}

	cfg := kvm.PITConfig{
		Flags: kvm.PITSpeakerDummy,
	}

	if err := kvm.CreatePIT2(vm, &cfg); err != nil {
		t.Fatal(err)
	}
}

func TestCPUID(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		t.Fatal(err)
	}

	defer sys.Close()

	ext, err := kvm.CheckExtension(sys, kvm.CapExtCPUID)
	if err != nil {
		t.Fatal(err)
	}

	if ext != 1 {
		t.Skipf("%v is %d", kvm.CapExtCPUID, ext)
	}

	ent, err := kvm.GetSupportedCPUID(sys)
	if err != nil {
		t.Fatal(err)
	}

	if len(ent) == 0 {
		t.Fatal("no supported cpuid")
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

	if err := kvm.SetCPUID2(vcpu, ent); err != nil {
		t.Fatal(err)
	}
}

func TestDeviceClosed_amd64(t *testing.T) {
	devFn := map[string]func(*os.File) error{
		"GetMSRIndexList":        func(sys *os.File) error { _, err := kvm.GetMSRIndexList(sys); return err },
		"GetMSRFeatureIndexList": func(sys *os.File) error { _, err := kvm.GetMSRFeatureIndexList(sys); return err },
		"GetSupportedCPUID":      func(sys *os.File) error { _, err := kvm.GetSupportedCPUID(sys); return err },
	}

	sys, err := os.Open("/dev/kvm")
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

func TestVMClosed_amd64(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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
		"CreateIRQChip":      func(vm *kvm.VM) error { return kvm.CreateIRQChip(vm) },
		"GetClock":           func(vm *kvm.VM) error { return kvm.GetClock(vm, nil) },
		"SetClock":           func(vm *kvm.VM) error { return kvm.SetClock(vm, nil) },
		"SetTSSAddr":         func(vm *kvm.VM) error { return kvm.SetTSSAddr(vm, 0) },
		"SetIdentityMapAddr": func(vm *kvm.VM) error { return kvm.SetIdentityMapAddr(vm, 0) },
		"CreatePIT2":         func(vm *kvm.VM) error { return kvm.CreatePIT2(vm, nil) },
	}

	for name, fn := range vmFn {
		if err := fn(vm); !errors.Is(err, unix.EBADF) {
			t.Fatalf("%s: %v != EBADF", name, err)
		}
	}
}

func TestVCPUClosed_amd64(t *testing.T) {
	sys, err := os.Open("/dev/kvm")
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

	vcpuFn := map[string]func(*kvm.VCPU) error{
		"GetRegs":   func(vcpu *kvm.VCPU) error { return kvm.GetRegs(vcpu, new(kvm.Regs)) },
		"SetRegs":   func(vcpu *kvm.VCPU) error { return kvm.SetRegs(vcpu, new(kvm.Regs)) },
		"GetSregs":  func(vcpu *kvm.VCPU) error { return kvm.GetSregs(vcpu, new(kvm.Sregs)) },
		"SetSregs":  func(vcpu *kvm.VCPU) error { return kvm.SetSregs(vcpu, new(kvm.Sregs)) },
		"GetMSRs":   func(vcpu *kvm.VCPU) error { _, err := kvm.GetMSRs(vcpu, nil); return err },
		"SetMSRs":   func(vcpu *kvm.VCPU) error { return kvm.SetMSRs(vcpu, nil) },
		"GetFPU":    func(vcpu *kvm.VCPU) error { return kvm.GetFPU(vcpu, new(kvm.FPU)) },
		"SetFPU":    func(vcpu *kvm.VCPU) error { return kvm.SetFPU(vcpu, new(kvm.FPU)) },
		"SetCPUID2": func(vcpu *kvm.VCPU) error { return kvm.SetCPUID2(vcpu, nil) },
	}

	for name, fn := range vcpuFn {
		if err := fn(vcpu); !errors.Is(err, unix.EBADF) {
			t.Fatalf("%s: %v != EBADF", name, err)
		}
	}
}
