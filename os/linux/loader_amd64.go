//go:build linux

package linux

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/c35s/hype/kvm"
	"github.com/c35s/hype/vmm"
	"github.com/c35s/hype/vmm/arch"
)

// Loader prepares the VM to boot a 64-bit Linux kernel in long mode.
type Loader struct {

	// Kernel is a bzImage.
	Kernel io.ReaderAt

	// Initrd, if set, is a compressed cpio of the initial ramdisk.
	Initrd io.Reader

	// Cmdline is the kernel command line.
	Cmdline string
}

const (
	gdtAddr      = 0x000001000
	pt4Addr      = 0x000002000
	pt3Addr      = 0x000003000
	pt2Addr      = 0x000004000
	zeropageAddr = 0x000010000
	cmdlineAddr  = 0x000020000
	kernelAddr   = 0x000100000
)

var gdt = []uint64{
	0, // NULL

	// FIX: derive entries from their corresponding kvm.Segment instead
	gdtEntry(0xa09b, 0, 0xfffff), // cs
	gdtEntry(0xc093, 0, 0xfffff), // ds
	gdtEntry(0x808b, 0, 0xfffff), // tss
}

func gdtEntry(flags uint16, base uint32, limit uint32) uint64 {
	return ((uint64(base) & 0xff000000) << (56 - 24)) |
		((uint64(flags) & 0x0000f0ff) << 40) |
		((uint64(limit) & 0x000f0000) << (48 - 16)) |
		((uint64(base) & 0x00ffffff) << 16) |
		(uint64(limit) & 0x0000ffff)
}

// loadflags

const (
	loadedHigh   = 1 << 0 // protected-mode code is loaded at 0x100000
	kaslrFlag    = 1 << 1 // used internally, KASLR enabled
	quietFlag    = 1 << 5 // suppress early messages
	keepSegments = 1 << 6 // do not reload the segment registers
	canUseHeap   = 1 << 7 // heap_end_ptr is valid
)

var le = binary.LittleEndian

func (l *Loader) LoadMemory(info vmm.VMInfo, mem []byte) error {
	var in BootParams

	zpg := make([]byte, 0x1000)
	if _, err := l.Kernel.ReadAt(zpg, 0x0); err != nil {
		panic(err)
	}

	if err := in.UnmarshalBinary(zpg); err != nil {
		panic(err)
	}

	if in.Hdr.Header != SetupHeaderMagic {
		panic("bzImage header doesn't have the magic")
	}

	if in.Hdr.Xloadflags&0b1 == 0 {
		panic("bzImage kernel doesn't have a 64-bit entrypoint at 0x200")
	}

	// build a clean zeropage
	params := BootParams{
		Hdr: in.Hdr,
	}

	// set obligatory fields
	params.Hdr.VidMode = 0xffff
	params.Hdr.TypeOfLoader = 0xff
	params.Hdr.Loadflags = loadedHigh

	// load the gdt
	for i, e := range gdt {
		le.PutUint64(mem[gdtAddr+i*8:], e)
	}

	// load page tables
	le.PutUint64(mem[pt4Addr:], pt3Addr|0x03)
	le.PutUint64(mem[pt3Addr:], pt2Addr|0x03)

	for i := uint64(0); i < 512; i++ {
		le.PutUint64(mem[pt2Addr+i*8:], (i<<21)+0x83)
	}

	// kernel cmdline
	var kargs []string

	// virtio-mmio devices
	for _, di := range info.Devices {
		kargs = append(kargs, fmt.Sprintf("virtio_mmio.device=%#x@%#x:%d", di.Size, di.Addr, di.IRQ))
	}

	// append the configured cmdline
	kargs = append(kargs, strings.Fields(l.Cmdline)...)
	cmdline := strings.Join(kargs, " ")

	// load the cmdline ASCIIZ
	copy(mem[cmdlineAddr:], append([]byte(cmdline), 0))

	params.Hdr.CmdLinePtr = cmdlineAddr
	params.Hdr.CmdlineSize = uint32(len(cmdline) + 1)

	if l.Initrd != nil {
		initrd, err := io.ReadAll(l.Initrd)
		if err != nil {
			panic(err)
		}

		// place initrd as high as possible
		initrdAddrMax := int(in.Hdr.InitrdAddrMax)

		// ...but no higher
		if initrdAddrMax > cap(mem) {
			initrdAddrMax = cap(mem)
		}

		initrdAddr := initrdAddrMax - len(initrd)

		// initrd must be page-aligned
		initrdAddr -= initrdAddr % 0x1000

		// load the initrd
		copy(mem[initrdAddr:], initrd)

		params.Hdr.RamdiskImage = uint32(initrdAddr)
		params.Hdr.RamdiskSize = uint32(len(initrd))
	}

	// set up the BIOS memory map
	// https://wiki.osdev.org/Memory_Map_(x86)
	// https://en.wikipedia.org/wiki/PCI_hole

	e820 := []BootE820Entry{
		{0x0, 0x0009fc00, 1}, // < 640K
	}

	switch {
	case cap(mem) < arch.MMIOHoleAddr:
		e820 = append(e820,
			BootE820Entry{kernelAddr, uint64(cap(mem)) - kernelAddr, 1})

	default:
		e820 = append(e820,
			BootE820Entry{kernelAddr, arch.MMIOHoleAddr - kernelAddr, 1},
			BootE820Entry{arch.AfterMMIOHoleAddr, uint64(cap(mem)) - arch.AfterMMIOHoleAddr, 1})
	}

	for i, e := range e820 {
		params.E820Table[i] = e
		params.E820Entries++
	}

	zeropage, err := params.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// load the zeropage
	copy(mem[zeropageAddr:], zeropage)

	// offset of the protected-mode kernel in the bzImage
	koff := (1 + int64(in.Hdr.SetupSects)) * 512

	// length of the protected-mode kernel
	klen := int(in.Hdr.Syssize) * 16

	if memsz, minsz := cap(mem), kernelAddr+klen; memsz < minsz {
		return errors.New("can't load kernel: guest memory is too small")
	}

	// load the protected-mode kernel
	if _, err := l.Kernel.ReadAt(mem[kernelAddr:kernelAddr+klen], koff); err != nil {
		return fmt.Errorf("can't load kernel: %w", err)
	}

	return nil
}

func (l *Loader) LoadVCPU(info vmm.VMInfo, slot int, regs *kvm.Regs, sregs *kvm.Sregs) error {
	if slot != 0 {
		panic("slot != 0")
	}

	sregs.GDT.Base = gdtAddr
	sregs.GDT.Limit = uint16(len(gdt)*8) - 1

	var cs = kvm.Segment{
		Base:     0x0,
		Limit:    0xfffff,
		Selector: 0x8,
		Type:     0xb,
		Present:  0x1,
		S:        0x1,
		L:        0x1,
		G:        0x1,
	}

	var ds = kvm.Segment{
		Base:     0x0,
		Limit:    0xfffff,
		Selector: 0x10,
		Type:     0x3,
		Present:  0x1,
		DB:       0x1,
		S:        0x1,
		G:        0x1,
	}

	var tss = kvm.Segment{
		Base:     0x0,
		Limit:    0xfffff,
		Selector: 0x18,
		Type:     0xb,
		Present:  0x1,
		G:        0x1,
	}

	sregs.CS = cs
	sregs.DS = ds
	sregs.ES = ds
	sregs.FS = ds
	sregs.GS = ds
	sregs.SS = ds
	sregs.TR = tss

	const (
		cr0PE   = 1 << 0
		cr0PG   = 1 << 31
		cr4PAE  = 1 << 5
		eferLME = 1 << 8
		eferLMA = 1 << 10
	)

	sregs.CR0 |= cr0PE | cr0PG
	sregs.CR3 = pt4Addr
	sregs.CR4 = cr4PAE
	sregs.EFER |= eferLME | eferLMA

	// intel arch reqt
	// FIX: isn't this already set?
	regs.RFlags = 0x2

	// addr of BootParams
	regs.RSI = zeropageAddr

	// 64-bit kernel entry point
	regs.RIP = kernelAddr + 0x200

	return nil
}
