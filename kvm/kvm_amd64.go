//go:build linux

package kvm

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Regs holds a VCPU's general-purpose registers.
// It has the same layout as the C struct kvm_regs.
type Regs struct {
	RAX, RBX, RCX, RDX uint64
	RSI, RDI, RSP, RBP uint64
	R8, R9, R10, R11   uint64
	R12, R13, R14, R15 uint64
	RIP, RFlags        uint64
}

// Sregs holds a VCPU's special registers.
// It has the same layout as the C struct kvm_sregs.
type Sregs struct {
	CS, DS, ES, FS, GS, SS  Segment
	TR, LDT                 Segment
	GDT, IDT                Dtable
	CR0, CR2, CR3, CR4, CR8 uint64
	EFER                    uint64
	APICBase                uint64
	InterruptBitmap         [((nrInterrupts + 63) / 64)]uint64
}

// Segment has the same layout as the C struct kvm_segment.
type Segment struct {
	Base                           uint64
	Limit                          uint32
	Selector                       uint16
	Type                           uint8
	Present, DPL, DB, S, L, G, Avl uint8
	Unusable                       uint8
	_                              byte
}

// Dtable has the same layout as the C struct kvm_dtable.
type Dtable struct {
	Base  uint64
	Limit uint16
	_     [6]byte
}

// MSREntry has the same layout as the C struct kvm_msr_entry.
type MSREntry struct {
	Index uint32
	_     uint32
	Data  uint64
}

// FPU holds a VCPU's floating-point state.
// It has the same layout as the C struct kvm_fpu.
type FPU struct {
	FPR        [8][16]byte
	FCW        uint16
	FSW        uint16
	FTWX       uint8
	_          uint8
	LastOpcode uint16
	LastIP     uint64
	LastDP     uint64
	XMM        [16][16]byte
	MXCSR      uint32
	_          uint32
}

// ClockData has the same layout as the C struct kvm_clock_data.
type ClockData struct {
	Clock uint64
	Flags uint32
	_     [36]byte
}

// CPUIDEntry has the same layout as the C struct kvm_cpuid_entry2.
type CPUIDEntry2 struct {
	Function uint32
	Index    uint32
	Flags    uint32
	EAX      uint32
	EBX      uint32
	ECX      uint32
	EDX      uint32
	_        [3]uint32
}

// PITConfig has the same layout as the C struct kvm_pit_config.
type PITConfig struct {
	Flags uint32
	_     [15]uint32
}

// VCPUState has roughly the same layout as struct kvm_run.
type VCPUState struct {
	_/*requestInterruptWindow*/ uint8 // in
	ImmediateExit                     uint8 // in
	_                                 [6]uint8
	ExitReason                        Exit
	_/*readyForInterruptInjection*/ uint8
	_/*ifFlag*/ uint8
	_/*flags*/ uint16
	_/*cr8*/ uint64
	_/*apicBase*/ uint64

	// exitData is a union of anonymous structs in the C struct.
	exitData [256]uint8

	_/*kvmValidRegs*/ uint64
	_/*kvmDirtyRegs*/ uint64
	_ [2048]uint8
}

// IOExitData is the result of a KVM_EXIT_IO vmexit. It has the same layout as the "io"
// member of the union of vmexit data in struct kvm_run.
type IOExitData struct {
	IsOut  bool
	Size   uint8
	Port   uint16
	Count  uint32
	Offset uint64
}

// MMIOExitData is the result of a KVM_EXIT_MMIO vmexit. It has the same layout as the
// "mmio" member of the union of vmexit data in struct kvm_run.
type MMIOExitData struct {
	PhysAddr uint64
	Data     [8]uint8
	Len      uint32
	IsWrite  bool
	_        [3]byte
}

// kvm_msr_list is similar to the C struct kvm_msr_list, which is used by the
// KVM_GET_MSR_INDEX_LIST and KVM_GET_MSR_FEATURE_INDEX_LIST ioctls. The indices array has
// a fixed size because Go doesn't directly support C flexible array members.
type kvm_msr_list struct {
	nmsrs   uint32
	indices [255]uint32
}

// kvm_msrs is similar to the C struct kvm_msrs, which is used by the KVM_GET_MSRS and
// KVM_SET_MSRS ioctls. The entries array has a fixed size because Go doesn't directly
// support C flexible array members.
type kvm_msrs struct {
	nmsrs   uint32
	entries [255]MSREntry
	_       uint32
}

// kvm_cpuid2 is similar to the C struct kvm_cpuid2.
type kvm_cpuid2 struct {
	nent    uint32
	_       uint32
	entries [255]CPUIDEntry2
}

// GetMSRIndexList "returns the guest msrs that are supported. The list
// varies by kvm version and host processor, but does not change otherwise."
func GetMSRIndexList(sys *os.File) (indices []int, err error) {
	var l kvm_msr_list
	l.nmsrs = uint32(len(l.indices))

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, sys.Fd(), kGetMSRIndexList, uintptr(unsafe.Pointer(&l)))
	if errno != 0 {
		return nil, errno
	}

	indices = make([]int, l.nmsrs)
	for i := range indices {
		indices[i] = int(l.indices[i])
	}

	return
}

// GetMSRFeatureIndexList "returns the list of MSRs that can be passed to the KVM_GET_MSRS
// system ioctl.  This lets userspace probe host capabilities and processor features that
// are exposed via MSRs (e.g., VMX capabilities). This list also varies by kvm version and
// host processor, but does not change otherwise."
//
// This ioctl is available if CheckExtension(CapGetMSRFeatures) returns 1.
func GetMSRFeatureIndexList(sys *os.File) (indices []int, err error) {
	var l kvm_msr_list
	l.nmsrs = uint32(len(l.indices))

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, sys.Fd(), kGetMSRFeatureIndexList, uintptr(unsafe.Pointer(&l)))
	if errno != 0 {
		return nil, errno
	}

	indices = make([]int, l.nmsrs)
	for i := range indices {
		indices[i] = int(l.indices[i])
	}

	return
}

// GetSupportedCPUID "returns x86 cpuid features which are supported by both the hardware
// and kvm in its default configuration. Userspace can use the information returned by
// this ioctl to construct cpuid information (for KVM_SET_CPUID2) that is consistent with
// hardware, kernel, and userspace capabilities, and with user requirements (for example,
// the user may wish to constrain cpuid to emulate older hardware, or for feature
// consistency across a cluster)."
//
// See 4.46 KVM_GET_SUPPORTED_CPUID in kvm/api.txt for more.
//
// This ioctl is available if CheckExtension(CapExtCPUID) returns 1.
func GetSupportedCPUID(sys *os.File) ([]CPUIDEntry2, error) {
	var cpuid kvm_cpuid2
	cpuid.nent = uint32(len(cpuid.entries))

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, sys.Fd(), kGetSupportedCPUID, uintptr(unsafe.Pointer(&cpuid)))
	if errno != 0 {
		return nil, errno
	}

	return cpuid.entries[:cpuid.nent], nil
}

// GetRegs reads the vcpu's general-purpose registers.
func GetRegs(vcpu *VCPU, regs *Regs) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kGetRegs, uintptr(unsafe.Pointer(regs)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetRegs writes the vcpu's general-purpose registers.
func SetRegs(vcpu *VCPU, regs *Regs) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kSetRegs, uintptr(unsafe.Pointer(regs)))
	if errno != 0 {
		return errno
	}

	return nil
}

// GetSregs reads the vcpu's special registers.
func GetSregs(vcpu *VCPU, sregs *Sregs) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kGetSregs, uintptr(unsafe.Pointer(sregs)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetSregs writes the vcpu's special registers.
func SetSregs(vcpu *VCPU, sregs *Sregs) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kSetSregs, uintptr(unsafe.Pointer(sregs)))
	if errno != 0 {
		return errno
	}

	return nil
}

// GetMSRs reads model-specific registers from the VCPU. The given indices should come from
// GetMSRIndexList. If CheckExtension(CapSetMSRFeatures) returns 1, GetMSRs can also read
// the values of MSR-based features that are available from the system. In this case, the
// given indices should come from GetMSRFeatureIndexList.
func GetMSRs(f interface{ Fd() uintptr }, indices []int) ([]MSREntry, error) {
	msrs := kvm_msrs{nmsrs: uint32(len(indices))}
	for i, index := range indices {
		msrs.entries[i].Index = uint32(index)
	}

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), kGetMSRs, uintptr(unsafe.Pointer(&msrs)))
	if errno != 0 {
		return nil, errno
	}

	return msrs.entries[:msrs.nmsrs], nil
}

// SetMSRs writes model-specific registers to the VCPU.
func SetMSRs(vcpu *VCPU, entries []MSREntry) error {
	msrs := kvm_msrs{nmsrs: uint32(len(entries))}
	if copy(msrs.entries[:], entries) != len(entries) {
		return unix.E2BIG
	}

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, vcpu.Fd(), kSetMSRs, uintptr(unsafe.Pointer(&msrs)))
	if errno != 0 {
		return errno
	}

	return nil
}

// GetFPU reads floating-point state from the VCPU.
func GetFPU(vcpu *VCPU, fpu *FPU) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kGetFPU, uintptr(unsafe.Pointer(fpu)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetFPU writes floating-point state to the VCPU.
func SetFPU(vcpu *VCPU, fpu *FPU) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vcpu.Fd(), kSetFPU, uintptr(unsafe.Pointer(fpu)))
	if errno != 0 {
		return errno
	}

	return nil
}

// GetClock returns "the current timestamp of kvmclock as seen by the current guest." This
// ioctl is available if CheckExtension(CapAdjustClock) returns a non-zero value.
func GetClock(vm *VM, data *ClockData) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vm.Fd(), kGetClock, uintptr(unsafe.Pointer(data)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetClock "[s]ets the current timestamp of kvmclock to the value specified in its
// parameter." This ioctl is available if CheckExtension(CapAdjustClock) returns a
// non-zero value.
func SetClock(vm *VM, data *ClockData) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vm.Fd(), kSetClock, uintptr(unsafe.Pointer(data)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetTSSAddr "defines the physical address of a three-page region in the guest physical
// address space. The region must be within the first 4GB of the guest physical address
// space and must not conflict with any memory slot or any mmio address. The guest may
// malfunction if it accesses this memory region."
//
// "This ioctl is required on Intel-based hosts. This is needed on Intel hardware because
// of a quirk in the virtualization implementation (see the internals documentation when
// it pops into existence)."
//
// This ioctl is available if CheckExtension(CapSetTSSAddr) returns 1.
func SetTSSAddr(vm *VM, addr uint64) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vm.Fd(), kSetTSSAddr, uintptr(addr))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetIdentityMapAddr "defines the physical address of a one-page region in the guest
// physical address space. The region must be within the first 4GB of the guest physical
// address space and must not conflict with any memory slot or any mmio address. The guest
// may malfunction if it accesses this memory region."
//
// "Setting the address to 0 will result in resetting the address to its default
// (0xfffbc000)."
//
// "This ioctl is required on Intel-based hosts. This is needed on Intel hardware because
// of a quirk in the virtualization implementation (see the internals documentation when
// it pops into existence)."
//
// SetIdentityMapAddr fails if it is called after CreateVCPU.
func SetIdentityMapAddr(vm *VM, addr uint64) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vm.Fd(), kSetIdentityMapAddr, uintptr(unsafe.Pointer(&addr)))
	if errno != 0 {
		return errno
	}

	return nil
}

// CreatePIT2 "Creates an in-kernel device model for the i8254 PIT. This call is only valid
// after enabling in-kernel irqchip support via KVM_CREATE_IRQCHIP."
//
// This ioctl is available if CheckExtension(CapPIT2) returns 1.
func CreatePIT2(vm *VM, cfg *PITConfig) error {
	_, _, errno := unix.Syscall(syscall.SYS_IOCTL, vm.Fd(), kCreatePIT2, uintptr(unsafe.Pointer(cfg)))
	if errno != 0 {
		return errno
	}

	return nil
}

// SetCPUID2 "defines the vcpu responses to the cpuid instruction."
// This ioctl is available if CheckExtension(CapExtCPUID) returns 1.
func SetCPUID2(vcpu *VCPU, entries []CPUIDEntry2) error {
	cpuid := kvm_cpuid2{nent: uint32(len(entries))}
	if copy(cpuid.entries[:], entries) != len(entries) {
		return unix.E2BIG
	}

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, vcpu.Fd(), kSetCPUID2, uintptr(unsafe.Pointer(&cpuid)))
	if errno != 0 {
		return errno
	}

	return nil
}

// IOExitData returns data describing the present KVM_EXIT_IO vmexit.
// The result is undefined (but bad) if the exit reason is not KVM_EXIT_IO.
func (s *VCPUState) IOExitData() *IOExitData {
	return (*IOExitData)(unsafe.Pointer(&s.exitData[0]))
}

// MMIOExitData returns data describing the present KVM_EXIT_MMIO vmexit.
// The result is undefined (but bad) if the exit reason is not KVM_EXIT_MMIO.
func (s *VCPUState) MMIOExitData() *MMIOExitData {
	return (*MMIOExitData)(unsafe.Pointer(&s.exitData[0]))
}
