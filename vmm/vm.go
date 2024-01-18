//go:build linux

// Package vmm provides helpers for configuring and running a KVM virtual machine.
package vmm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/virtio/mmio"
	"github.com/c35s/hype/vmm/arch"
	"golang.org/x/sys/unix"
)

// Config describes a new VM.
type Config struct {

	// MemSize is the size of the VM's memory in bytes.
	// It must be a multiple of the host's page size.
	// If MemSize is 0, the VM will have 1G of memory.
	MemSize int

	// Devices configures the VM's virtio-mmio devices.
	Devices []virtio.DeviceHandler

	// Loader configures the VM's memory and registers.
	Loader Loader

	// Arch, if set, is called to do arch-specific setup during VM creation.
	// If Arch is nil, a default implementation is used. Setting Arch is
	// probably only useful for testing, debugging, and development.
	Arch Arch
}

// VMInfo describes a configured VM in a form useful to the Loader.
// It is passed to the Loader's LoadMemory and LoadVCPU methods.
type VMInfo struct {

	// MemSize is the size of the VM's memory in bytes.
	// It is a multiple of the host's page size.
	MemSize int

	// NumCPU is the number of VCPUs attached to the VM.
	// Right now it's always 1.
	NumCPU int

	// Devices enumerates the VM's virtio-mmio devices.
	Devices []mmio.DeviceInfo
}

type Loader interface {

	// LoadMemory prepares the VM's memory before it boots.
	LoadMemory(info VMInfo, mem []byte) error

	// LoadVCPU prepares a VCPU before the VM boots.
	LoadVCPU(info VMInfo, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error
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

type VM struct {
	fd   *kvm.VM
	mem  []byte
	cpu  []*vcpu
	mmio *mmio.Bus
	irqf map[int]int // irq:fd
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

// vcpu collects a VCPU fd and its mmaped state.
type vcpu struct {
	fd *kvm.VCPU
	mm []byte
}

// New creates a new VM.
func New(cfg Config) (*VM, error) {
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
	cpu := make([]*vcpu, 1)
	for slot := range cpu {
		fd, err := kvm.CreateVCPU(vm, slot)
		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrCreateVCPU, slot, err)
		}

		mm, err := unix.Mmap(int(fd.Fd()), 0, mmsz,
			unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrMmapVCPU, slot, err)
		}

		cpu[slot] = &vcpu{fd: fd, mm: mm}
		if err := cfg.Arch.SetupVCPU(slot, cpu[slot].fd, cpu[slot].State()); err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrSetupVCPU, slot, err)
		}
	}

	m := &VM{
		fd:   vm,
		cpu:  cpu,
		mem:  mem,
		irqf: make(map[int]int),
	}

	// configure the virtio-mmio bus to call back to the VM
	m.mmio = mmio.NewBus(cfg.Devices, m.mmioMemAt, m.mmioNotify)

	info := VMInfo{
		MemSize: len(m.mem),
		NumCPU:  len(m.cpu),
		Devices: m.mmio.Devices(),
	}

	// wire up device irqs
	for _, di := range info.Devices {
		fd, err := unix.Eventfd(0, unix.EFD_CLOEXEC)
		if err != nil {
			panic(err)
		}

		err = kvm.IRQFD(m.fd, &kvm.IRQFDConfig{
			Fd:  uint32(fd),
			GSI: uint32(di.IRQ),
		})

		if err != nil {
			panic(err)
		}

		m.irqf[di.IRQ] = fd
	}

	// load memory
	if err := cfg.Loader.LoadMemory(info, m.mem); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadMemory, err)
	}

	// load VCPUs
	for slot, c := range m.cpu {
		err := func() error {
			var (
				regs  kvm.Regs
				sregs kvm.Sregs
			)

			if err := kvm.GetRegs(c.fd, &regs); err != nil {
				return fmt.Errorf("get regs: %w", err)
			}

			if err := kvm.GetSregs(c.fd, &sregs); err != nil {
				return fmt.Errorf("get sregs: %w", err)
			}

			if err := cfg.Loader.LoadVCPU(info, slot, &regs, &sregs); err != nil {
				return err
			}

			if err := kvm.SetRegs(c.fd, &regs); err != nil {
				return fmt.Errorf("set regs: %w", err)
			}

			if err := kvm.SetSregs(c.fd, &sregs); err != nil {
				return fmt.Errorf("set sregs: %w", err)
			}

			return nil
		}()

		if err != nil {
			return nil, fmt.Errorf("%w: slot %d: %w", ErrLoadVCPU, slot, err)
		}
	}

	return m, nil
}

func (m *VM) Run(context.Context) error {
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
		case kvm.ExitIO:
			continue

		case kvm.ExitMMIO:
			xd := state.MMIOExitData()
			if _, err := m.mmio.HandleMMIO(xd.PhysAddr, xd.Data[:xd.Len], xd.IsWrite); err != nil {
				panic(err)
			}

		case kvm.ExitShutdown:
			return nil

		default:
			panic(reason)
		}
	}
}

func (m *VM) Close() error {
	for _, c := range m.cpu {
		c.Close()
	}

	m.fd.Close()
	unix.Munmap(m.mem)
	m.mem = nil

	return nil
}

func (m *VM) mmioMemAt(addr uint64, len int) ([]byte, error) {
	return m.mem[addr : addr+uint64(len)], nil
}

func (m *VM) mmioNotify(irq int) error {
	if fd, ok := m.irqf[irq]; ok {
		if _, err := unix.Write(fd, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
			return err
		}
	}

	return nil
}

func (c *vcpu) State() *kvm.VCPUState {
	return (*kvm.VCPUState)(unsafe.Pointer(&c.mm[0]))
}

func (c *vcpu) Close() error {
	c.fd.Close()
	unix.Munmap(c.mm)
	return nil
}

func (cfg Config) validate() error {
	if pgsz := os.Getpagesize(); cfg.MemSize%pgsz != 0 {
		return fmt.Errorf("memory size must be a multiple of the host page size (%d)", pgsz)
	}

	if cfg.MemSize < MemSizeMin {
		return fmt.Errorf("memory is too small: %d < %d", cfg.MemSize, MemSizeMin)
	}

	if cfg.MemSize > MemSizeMax {
		return fmt.Errorf("memory is too large: %d > %d", cfg.MemSize, MemSizeMax)
	}

	if cfg.Loader == nil {
		return errors.New("loader is not set")
	}

	return nil
}

func (cfg Config) withDefaults() Config {
	if cfg.MemSize == 0 {
		cfg.MemSize = MemSizeDefault
	}

	return cfg
}
