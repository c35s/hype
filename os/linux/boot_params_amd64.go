//go:build linux && amd64

package linux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// BootParams is the so-called "zeropage". It corresponds to struct boot_params. Since
// struct boot_params is packed, BootParams doesn't have exactly the same layout. Instead,
// it implements BinaryUnmarshaler and BinaryMarshaler.
type BootParams struct {
	_           [64]byte   // 	struct screen_info screen_info;			/* 0x000 */
	_           [20]byte   // 	struct apm_bios_info apm_bios_info;		/* 0x040 */
	_           [4]uint8   // 	__u8  _pad2[4];					/* 0x054 */
	_           uint64     // 	__u64  tboot_addr;				/* 0x058 */
	_           [16]byte   // 	struct ist_info ist_info;			/* 0x060 */
	_           uint64     // 	__u64 acpi_rsdp_addr;				/* 0x070 */
	_           [8]uint8   // 	__u8  _pad3[8];					/* 0x078 */
	_           [16]uint8  // 	__u8  hd0_info[16];	/* obsolete! */		/* 0x080 */
	_           [16]uint8  // 	__u8  hd1_info[16];	/* obsolete! */		/* 0x090 */
	_           [16]byte   // 	struct sys_desc_table sys_desc_table; /* obsolete! */	/* 0x0a0 */
	_           [16]byte   // 	struct olpc_ofw_header olpc_ofw_header;		/* 0x0b0 */
	_           uint32     // 	__u32 ext_ramdisk_image;			/* 0x0c0 */
	_           uint32     // 	__u32 ext_ramdisk_size;				/* 0x0c4 */
	_           uint32     // 	__u32 ext_cmd_line_ptr;				/* 0x0c8 */
	_           [112]uint8 // 	__u8  _pad4[112];				/* 0x0cc */
	_           uint32     // 	__u32 cc_blob_address;				/* 0x13c */
	_           [128]byte  // 	struct edid_info edid_info;			/* 0x140 */
	_           [32]byte   // 	struct efi_info efi_info;			/* 0x1c0 */
	_           uint32     // 	__u32 alt_mem_k;				/* 0x1e0 */
	_           uint32     // 	__u32 scratch;		/* Scratch field! */	/* 0x1e4 */
	E820Entries uint8      // 	__u8  e820_entries;				/* 0x1e8 */
	_           uint8      // 	__u8  eddbuf_entries;				/* 0x1e9 */
	_           uint8      // 	__u8  edd_mbr_sig_buf_entries;			/* 0x1ea */
	_           uint8      // 	__u8  kbd_status;				/* 0x1eb */
	_           uint8      // 	__u8  secure_boot;				/* 0x1ec */
	_           [2]uint8   // 	__u8  _pad5[2];					/* 0x1ed */

	// 	/*
	// 	 * The sentinel is set to a nonzero value (0xff) in header.S.
	// 	 *
	// 	 * A bootloader is supposed to only take setup_header and put
	// 	 * it into a clean boot_params buffer. If it turns out that
	// 	 * it is clumsy or too generous with the buffer, it most
	// 	 * probably will pick up the sentinel variable too. The fact
	// 	 * that this variable then is still 0xff will let kernel
	// 	 * know that some variables in boot_params are invalid and
	// 	 * kernel should zero out certain portions of boot_params.
	// 	 */
	// 	__u8  sentinel;					/* 0x1ef */
	_ uint8

	_         uint8              // 	__u8  _pad6[1];					/* 0x1f0 */
	Hdr       SetupHeader        // 	struct setup_header hdr;    /* setup header */	/* 0x1f1 */
	_         [36]uint8          // 	__u8  _pad7[0x290-0x1f1-sizeof(struct setup_header)];
	_         [64]byte           // 	__u32 edd_mbr_sig_buffer[EDD_MBR_SIG_MAX];	/* 0x290 */
	E820Table [128]BootE820Entry // 	struct boot_e820_entry e820_table[E820_MAX_ENTRIES_ZEROPAGE]; /* 0x2d0 */
	_         [48]uint8          // 	__u8  _pad8[48];				/* 0xcd0 */
	_         [492]byte          // 	struct edd_info eddbuf[EDDMAXNR];		/* 0xd00 */
	_         [276]uint8         // 	__u8  _pad9[276];				/* 0xeec */
}

// bootE820Entry represents the E820 memory region entry of the boot protocol ABI. It
// corresponds to struct boot_e820_entry. But since struct boot_e820_entry is packed, they
// don't have exactly the same layout.
type BootE820Entry struct {
	Addr uint64 // __u64 addr
	Size uint64 // __u64 size
	Type uint32 // __u32 type
}

// SetupHeader is the part of the zeropage that explains how to boot the kernel. A boot
// loader usually copies the SetupHeader from of the kernel image's BootParams, customizes
// it, and copies it to the zeropage in memory. SetupHeader corresponds to struct
// setup_header, but they don't have exactly the same layout because the C struct is
// packed.
type SetupHeader struct {
	SetupSects          uint8  // __u8 setup_sects
	RootFlags           uint16 // __u16 root_flags
	Syssize             uint32 // __u32 syssize
	RamSize             uint16 // __u16 ram_size
	VidMode             uint16 // __u16 vid_mode
	RootDev             uint16 // __u16 root_dev
	BootFlag            uint16 // __u16 boot_flag
	Jump                uint16 // __u16 jump
	Header              uint32 // __u32 header
	Version             uint16 // __u16 version
	RealmodeSwtch       uint32 // __u32 realmode_swtch
	StartSysSeg         uint16 // __u16 start_sys_seg
	KernelVersion       uint16 // __u16 kernel_version
	TypeOfLoader        uint8  // __u8 type_of_loader
	Loadflags           uint8  // __u8 loadflags
	SetupMoveSize       uint16 // __u16 setup_move_size
	Code32Start         uint32 // __u32 code32_start
	RamdiskImage        uint32 // __u32 ramdisk_image
	RamdiskSize         uint32 // __u32 ramdisk_size
	BootsectKludge      uint32 // __u32 bootsect_kludge
	HeapEndPtr          uint16 // __u16 heap_end_ptr
	ExtLoaderVer        uint8  // __u8 ext_loader_ver
	ExtLoaderType       uint8  // __u8 ext_loader_type
	CmdLinePtr          uint32 // __u32 cmd_line_ptr
	InitrdAddrMax       uint32 // __u32 initrd_addr_max
	KernelAlignment     uint32 // __u32 kernel_alignment
	RelocatableKernel   uint8  // __u8 relocatable_kernel
	MinAlignment        uint8  // __u8 min_alignment
	Xloadflags          uint16 // __u16 xloadflags
	CmdlineSize         uint32 // __u32 cmdline_size
	HardwareSubarch     uint32 // __u32 hardware_subarch
	HardwareSubarchData uint64 // __u64 hardware_subarch_data
	PayloadOffset       uint32 // __u32 payload_offset
	PayloadLength       uint32 // __u32 payload_length
	SetupData           uint64 // __u64 setup_data
	PrefAddress         uint64 // __u64 pref_address
	InitSize            uint32 // __u32 init_size
	HandoverOffset      uint32 // __u32 handover_offset
	KernelInfoOffset    uint32 // __u32 kernel_info_offset
}

// SetupHeaderMagic is the required value of the SetupHeader.Header field.
const SetupHeaderMagic = 0x53726448 // "HdrS"

// ZeropageSize is the size of the zeropage in bytes (4K).
const ZeropageSize = 0x1000

var ErrBzImageMagic = errors.New("linux: parse bzImage: bad header magic")

// ParseBzImage reads 4096 bytes from r at offset 0 and parses it into a BootParams struct.
func ParseBzImage(r io.ReaderAt) (zeropage *BootParams, err error) {
	z := make([]byte, ZeropageSize)
	if _, err := r.ReadAt(z, 0); err != nil {
		return nil, err
	}

	zeropage = new(BootParams)
	zeropage.UnmarshalBinary(z)

	if zeropage.Hdr.Header != SetupHeaderMagic {
		return nil, fmt.Errorf("%w: %#x != %#x", ErrBzImageMagic, zeropage.Hdr.Header, SetupHeaderMagic)
	}

	return
}

// MarshalBinary marshals the params into the layout of struct boot_params.
func (bp *BootParams) MarshalBinary() (data []byte, err error) {
	b := new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, bp); err != nil {
		panic(err)
	}

	return b.Bytes(), nil
}

// UnmarshalBinary unmarshals a packed struct boot_params into the params.
// It returns io.ErrUnexpectedEOF if the given data is too short.
func (bp *BootParams) UnmarshalBinary(data []byte) error {
	if len(data) < ZeropageSize {
		return io.ErrUnexpectedEOF
	}

	r := bytes.NewReader(data[:ZeropageSize])
	if err := binary.Read(r, binary.LittleEndian, bp); err != nil {
		panic(err)
	}

	return nil
}
