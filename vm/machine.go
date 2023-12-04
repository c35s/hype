//go:build linux

// Package vm provides helpers for configuring and running a KVM virtual machine.
package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/c35s/hype/kvm"
	"golang.org/x/sys/unix"
)

// Config describes a new VM.
type Config struct {

	// MemSize is the size of the VM's memory in bytes.
	// It must be a multiple of the host's page size.
	// If MemSize is 0, the VM will have 1G of memory.
	MemSize int

	// Loader configures the VM's memory and registers.
	Loader Loader
}

type Loader interface {

	// LoadMemory prepares the VM's memory before it boots.
	LoadMemory(mem []byte) error

	// LoadVCPU prepares a VCPU before the VM boots.
	LoadVCPU(slot int, regs *kvm.Regs, sregs *kvm.Sregs) error
}

type Machine struct {
	fd  *kvm.VM
	mem []byte
	cpu []proc
}

const (
	MemSizeMin     = 1 << 20 // 1M
	MemSizeDefault = 1 << 30 // 1G
	MemSizeMax     = 1 << 40 // 1T
)

var (
	ErrOpenKVM             = errors.New("vm: KVM is not available")
	ErrCompat              = errors.New("vm: incompatible KVM")
	ErrConfig              = errors.New("vm: invalid config")
	ErrGetVCPUMmapSize     = errors.New("vm: get VCPU mmap size failed")
	ErrCreate              = errors.New("vm: create failed")
	ErrAllocMemory         = errors.New("vm: memory allocation failed")
	ErrLoadMemory          = errors.New("vm: memory load failed")
	ErrSetUserMemoryRegion = errors.New("vm: set user memory region failed")
	ErrCreateVCPU          = errors.New("vm: VCPU create failed")
	ErrMmapVCPU            = errors.New("vm: VCPU mmap failed")
	ErrLoadVCPU            = errors.New("vm: VCPU load failed")
)

// proc collects a VCPU fd and its mmaped state.
type proc struct {
	fd *kvm.VCPU
	mm []byte
}

// New creates a new VM.
func New(cfg Config) (*Machine, error) {
	sys, err := kvm.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenKVM, err)
	}

	defer sys.Close()

	if err := testKVMCompat(sys); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCompat, err)
	}

	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfig, err)
	}

	mmsz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetVCPUMmapSize, err)
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreate, err)
	}

	mem, err := unix.Mmap(-1, 0, cfg.MemSize,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_NORESERVE)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAllocMemory, err)
	}

	if err := cfg.Loader.LoadMemory(mem); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadMemory, err)
	}

	reg := kvm.UserspaceMemoryRegion{
		MemorySize:    uint64(len(mem)),
		UserspaceAddr: uint64(uintptr(unsafe.Pointer(&mem[0]))),
	}

	if err := kvm.SetUserMemoryRegion(vm, &reg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSetUserMemoryRegion, err)
	}

	cpu := make([]proc, 1)
	for slot := range cpu {
		c, err := kvm.CreateVCPU(vm, slot)
		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrCreateVCPU, slot, err)
		}

		mm, err := unix.Mmap(int(c.Fd()), 0, mmsz,
			unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrMmapVCPU, slot, err)
		}

		loadErr := func() error {
			var (
				regs  kvm.Regs
				sregs kvm.Sregs
			)

			if err := kvm.GetRegs(c, &regs); err != nil {
				return fmt.Errorf("get regs: %w", err)
			}

			if err := kvm.GetSregs(c, &sregs); err != nil {
				return fmt.Errorf("get sregs: %w", err)
			}

			if err := cfg.Loader.LoadVCPU(slot, &regs, &sregs); err != nil {
				return err
			}

			if err := kvm.SetRegs(c, &regs); err != nil {
				return fmt.Errorf("set regs: %w", err)
			}

			if err := kvm.SetSregs(c, &sregs); err != nil {
				return fmt.Errorf("set sregs: %w", err)
			}

			return nil
		}()

		if loadErr != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrLoadVCPU, slot, loadErr)
		}

		cpu[slot] = proc{
			fd: c,
			mm: mm,
		}
	}

	m := Machine{
		fd:  vm,
		cpu: cpu,
		mem: mem,
	}

	return &m, nil
}

func (m *Machine) Run(context.Context) error {
	for {
		if err := kvm.Run(m.cpu[0].fd); err != nil {
			panic(err)
		}

		var (
			state  = m.cpu[0].State()
			reason = state.ExitReason
		)

		switch reason {
		case kvm.ExitHLT:
			return nil

		default:
			panic(reason)
		}
	}
}

func (m *Machine) Close() error {
	for _, p := range m.cpu {
		p.fd.Close()
		unix.Munmap(p.mm)
	}

	m.fd.Close()
	unix.Munmap(m.mem)
	m.mem = nil

	return nil
}

func (c Config) validate() error {
	if pgsz := os.Getpagesize(); c.MemSize%pgsz != 0 {
		return fmt.Errorf("memory size must be a multiple of the host page size (%d)", pgsz)
	}

	if c.MemSize < MemSizeMin {
		return fmt.Errorf("memory is too small: %d < %d", c.MemSize, MemSizeMin)
	}

	if c.MemSize > MemSizeMax {
		return fmt.Errorf("memory is too large: %d > %d", c.MemSize, MemSizeMax)
	}

	return nil
}

func (c Config) withDefaults() Config {
	if c.MemSize == 0 {
		c.MemSize = MemSizeDefault
	}

	return c
}

func (p *proc) State() *kvm.VCPUState {
	return (*kvm.VCPUState)(unsafe.Pointer(&p.mm[0]))
}

func testKVMCompat(sys *kvm.System) error {
	version, err := kvm.GetAPIVersion(sys)
	if err != nil {
		return err
	}

	if version != kvm.StableAPIVersion {
		return fmt.Errorf("unstable API version: %d != %d", version, kvm.StableAPIVersion)
	}

	// FIX: check require arch-specific ext somewhere else

	required := []kvm.Cap{
		kvm.CapIRQChip,
		kvm.CapHLT,
		kvm.CapUserMemory,
		kvm.CapCheckExtensionVM,
	}

	var missing []kvm.Cap
	for _, cap := range required {
		val, err := kvm.CheckExtension(sys, cap)
		if err != nil {
			return err
		}

		if val < 1 {
			missing = append(missing, cap)
		}
	}

	if len(missing) > 0 {
		var names []string
		for _, cap := range missing {
			names = append(names, cap.String())
		}

		return fmt.Errorf("missing %s", strings.Join(names, ","))
	}

	return nil
}
