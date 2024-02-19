//go:build linux

// Package vmm provides helpers for configuring and running a KVM virtual machine.
package vmm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
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
	Devices []virtio.DeviceConfig

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

	mu    sync.Mutex
	doneC chan struct{}
}

const (
	MemSizeMin     = 1 << 20 // 1M
	MemSizeDefault = 1 << 30 // 1G
	MemSizeMax     = 1 << 40 // 1T
)

var (
	ErrOpenKVM             = errors.New("vmm: KVM is not available")
	ErrCompat              = errors.New("vmm: incompatible KVM")
	ErrConfig              = errors.New("vmm: invalid config")
	ErrGetVCPUMmapSize     = errors.New("vmm: get VCPU mmap size failed")
	ErrCreate              = errors.New("vmm: create failed")
	ErrSetup               = errors.New("vmm: setup failed")
	ErrAllocMemory         = errors.New("vmm: memory allocation failed")
	ErrSetupMemory         = errors.New("vmm: memory setup failed")
	ErrLoadMemory          = errors.New("vmm: memory load failed")
	ErrSetUserMemoryRegion = errors.New("vmm: set user memory region failed")
	ErrCreateVCPU          = errors.New("vmm: VCPU create failed")
	ErrMmapVCPU            = errors.New("vmm: VCPU mmap failed")
	ErrSetupVCPU           = errors.New("vmm: VCPU setup failed")
	ErrLoadVCPU            = errors.New("vmm: VCPU load failed")
	ErrVMClosed            = errors.New("vmm: VM closed")
)

// vcpu collects a VCPU fd and its mmaped state.
type vcpu struct {
	fd    *kvm.VCPU
	mm    []byte
	opC   chan vcpuOp
	doneC chan struct{}
}

// vcpuOp is an operation to be performed on a vcpu thread.
type vcpuOp struct {
	F func() error
	C chan error
}

func init() {
	signal.Ignore(unix.SIGURG)
}

// New creates a new VM.
func New(cfg Config) (*VM, error) {
	sys, err := os.Open("/dev/kvm")
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
		c := &vcpu{
			opC:   make(chan vcpuOp),
			doneC: make(chan struct{}),
		}

		go func() {
			defer close(c.doneC)
			runtime.LockOSThread()
			for op := range c.opC {
				op.C <- op.F()
			}

			if c.fd != nil {
				c.fd.Close()
			}

			if c.mm != nil {
				unix.Munmap(c.mm)
			}
		}()

		err := c.Do(func() error {
			fd, err := kvm.CreateVCPU(vm, slot)
			if err != nil {
				return fmt.Errorf("%w: slot %d: %w", ErrCreateVCPU, slot, err)
			}
			c.fd = fd

			mm, err := unix.Mmap(int(fd.Fd()), 0, mmsz,
				unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)

			if err != nil {
				return fmt.Errorf("%w: slot %d: %w", ErrMmapVCPU, slot, err)
			}
			c.mm = mm

			if err := cfg.Arch.SetupVCPU(slot, c.fd, c.State()); err != nil {
				return fmt.Errorf("%w: slot %d: %w", ErrSetupVCPU, slot, err)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		cpu[slot] = c
	}

	m := &VM{
		fd:    vm,
		cpu:   cpu,
		mem:   mem,
		irqf:  make(map[int]int),
		doneC: make(chan struct{}),
	}

	m.mmio, err = mmio.NewBus(cfg.Devices, mmio.Config{
		MemAt: func(addr uint64, len int) ([]byte, error) {
			return m.mem[addr : addr+uint64(len)], nil
		},

		Notify: func(irq int) error {
			if fd, ok := m.irqf[irq]; ok {
				if _, err := unix.Write(fd, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
					return err
				}
			}

			return nil
		},
	})

	if err != nil {
		return nil, fmt.Errorf("vm: create mmio bus: %w", err)
	}

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

func (m *VM) Run(ctx context.Context) error {
	select {
	case <-m.doneC:
		return ErrVMClosed

	default:
		break
	}

	go func() {
		select {
		case <-m.doneC:
			return

		case <-ctx.Done():
			m.cpu[0].State().ImmediateExit = 1
		}
	}()

	return m.cpu[0].Do(func() error {
		for {
			if err := kvm.Run(m.cpu[0].fd); err != nil {
				if err == unix.EINTR {
					return ctx.Err()
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
	})
}

// Close stops the VM and releases its resources. It returns ErrVMClosed if the
// VM is already closed. Close closes the VCPUs and waits for them to stop. Then
// it closes the MMIO bus, which closes each of its devices in turn. Then the
// underlying VM fd is closed and the VM's memory is munmaped.
func (m *VM) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.doneC:
		return ErrVMClosed

	default:
		close(m.doneC)
	}

	// wait for the vcpus
	for _, c := range m.cpu {
		close(c.opC)
		<-c.doneC
	}

	if err := m.mmio.Close(); err != nil {
		return fmt.Errorf("close mmio: %w", err)
	}

	if err := m.fd.Close(); err != nil {
		return fmt.Errorf("close vm fd: %w", err)
	}

	if err := unix.Munmap(m.mem); err != nil {
		return fmt.Errorf("unmap memory: %w", err)
	}

	return nil
}

func (c *vcpu) State() *kvm.VCPUState {
	return (*kvm.VCPUState)(unsafe.Pointer(&c.mm[0]))
}

// Do runs f on the VCPU's thread and returns its result. Calls are serialized.
func (c *vcpu) Do(f func() error) error {
	op := vcpuOp{f, make(chan error)}
	c.opC <- op
	return <-op.C
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
