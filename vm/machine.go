//go:build linux

// Package vm provides helpers for configuring and running a KVM virtual machine.
package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vm/arch"
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

	// Arch, if set, is called to do arch-specific setup during VM creation.
	// If Arch is nil, a default implementation is used. Setting Arch is
	// probably only useful for testing, debugging, and development.
	Arch Arch
}

// Info describes a configured VM in a form useful to the Loader.
// It is passed to the Loader's LoadMemory and LoadVCPU methods.
type Info struct {

	// MemSize is the size of the VM's memory in bytes.
	// It is a multiple of the host's page size.
	MemSize int

	// NumCPU is the number of VCPUs attached to the VM.
	// Right now it's always 1.
	NumCPU int
}

type Loader interface {

	// LoadMemory prepares the VM's memory before it boots.
	LoadMemory(vm Info, mem []byte) error

	// LoadVCPU prepares a VCPU before the VM boots.
	LoadVCPU(vm Info, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error
}

type Arch interface {

	// SetupVM is called after the VM is created.
	// It sets up arch-specific "hardware" like the PIC.
	SetupVM(vm *kvm.VM) error

	// SetupMemory is called after the VM's memory is allocated.
	// It partitions the memory into regions. It can also write
	// arch-specific data to the memory if necessary.
	SetupMemory(mem []byte) ([]kvm.UserspaceMemoryRegion, error)

	// SetupVCPU is called after the VCPU is created and mmaped.
	// It sets up arch-specific features like MSRs and cpuid.
	SetupVCPU(slot int, vcpu *kvm.VCPU, state *kvm.VCPUState) error
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
	ErrSetup               = errors.New("vm: setup failed")
	ErrAllocMemory         = errors.New("vm: memory allocation failed")
	ErrSetupMemory         = errors.New("vm: memory setup failed")
	ErrLoadMemory          = errors.New("vm: memory load failed")
	ErrSetUserMemoryRegion = errors.New("vm: set user memory region failed")
	ErrCreateVCPU          = errors.New("vm: VCPU create failed")
	ErrMmapVCPU            = errors.New("vm: VCPU mmap failed")
	ErrSetupVCPU           = errors.New("vm: VCPU setup failed")
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

	if err := arch.ValidateKVM(sys); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCompat, err)
	}

	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfig, err)
	}

	// default arch
	if cfg.Arch == nil {
		a, err := arch.New(sys)
		if err != nil {
			panic(err)
		}

		cfg.Arch = a
	}

	vm, err := kvm.CreateVM(sys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreate, err)
	}

	// install arch-specific "hardware"
	if err := cfg.Arch.SetupVM(vm); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSetup, err)
	}

	// create memory
	mem, err := unix.Mmap(-1, 0, cfg.MemSize,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_NORESERVE)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAllocMemory, err)
	}

	// partition memory
	mrs, err := cfg.Arch.SetupMemory(mem)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSetupMemory, err)
	}

	// install memory
	for _, mr := range mrs {
		if err := kvm.SetUserMemoryRegion(vm, &mr); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSetUserMemoryRegion, err)
		}
	}

	mmsz, err := kvm.GetVCPUMmapSize(sys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetVCPUMmapSize, err)
	}

	// create VCPUs
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

		p := proc{fd: c, mm: mm}
		if err := cfg.Arch.SetupVCPU(slot, p.fd, p.State()); err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrSetupVCPU, slot, err)
		}

		cpu[slot] = p
	}

	info := Info{
		MemSize: len(mem),
		NumCPU:  len(cpu),
	}

	// load memory
	if err := cfg.Loader.LoadMemory(info, mem); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadMemory, err)
	}

	// load VCPUs
	for slot, p := range cpu {
		err := func() error {
			var (
				regs  kvm.Regs
				sregs kvm.Sregs
			)

			if err := kvm.GetRegs(p.fd, &regs); err != nil {
				return fmt.Errorf("get regs: %w", err)
			}

			if err := kvm.GetSregs(p.fd, &sregs); err != nil {
				return fmt.Errorf("get sregs: %w", err)
			}

			if err := cfg.Loader.LoadVCPU(info, slot, &regs, &sregs); err != nil {
				return err
			}

			if err := kvm.SetRegs(p.fd, &regs); err != nil {
				return fmt.Errorf("set regs: %w", err)
			}

			if err := kvm.SetSregs(p.fd, &sregs); err != nil {
				return fmt.Errorf("set sregs: %w", err)
			}

			return nil
		}()

		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrLoadVCPU, slot, err)
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
			if err == unix.EINTR {
				continue
			}

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

	if c.Loader == nil {
		return errors.New("loader is not set")
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
