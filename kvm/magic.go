// DO NOT EDIT! Generated by cmd/kvm-gen-magic.
//go:build linux

package kvm

import "fmt"

const (
	CapIRQChip                  = Cap(0)
	CapHLT                      = Cap(1)
	CapMMUShadowCacheControl    = Cap(2)
	CapUserMemory               = Cap(3)
	CapSetTSSAddr               = Cap(4)
	CapVAPIC                    = Cap(6)
	CapExtCPUID                 = Cap(7)
	CapClockSource              = Cap(8)
	CapNrVCPUs                  = Cap(9)
	CapNrMemslots               = Cap(10)
	CapPIT                      = Cap(11)
	CapNopIODelay               = Cap(12)
	CapPVMMU                    = Cap(13)
	CapMPState                  = Cap(14)
	CapCoalescedMMIO            = Cap(15)
	CapSyncMMU                  = Cap(16)
	CapIOMMU                    = Cap(18)
	CapDestroyMemoryRegionWorks = Cap(21)
	CapUserNMI                  = Cap(22)
	CapSetGuestDebug            = Cap(23)
	CapReinjectControl          = Cap(24)
	CapIRQRouting               = Cap(25)
	CapIRQInjectStatus          = Cap(26)
	CapAssignDevIRQ             = Cap(29)
	CapJoinMemoryRegionsWorks   = Cap(30)
	CapMCE                      = Cap(31)
	CapIRQFD                    = Cap(32)
	CapPIT2                     = Cap(33)
	CapSetBootCPUID             = Cap(34)
	CapPITState2                = Cap(35)
	CapIOEventFD                = Cap(36)
	CapSetIdentityMapAddr       = Cap(37)
	CapXenHVM                   = Cap(38)
	CapAdjustClock              = Cap(39)
	CapInternalErrorData        = Cap(40)
	CapVCPUEvents               = Cap(41)
	CapS390PSW                  = Cap(42)
	CapPPCSegState              = Cap(43)
	CapHyperV                   = Cap(44)
	CapHyperVVAPIC              = Cap(45)
	CapHyperVSpin               = Cap(46)
	CapPCISegment               = Cap(47)
	CapPPCPairedSingles         = Cap(48)
	CapIntrShadow               = Cap(49)
	CapDebugRegs                = Cap(50)
	CapX86RobustSingleStep      = Cap(51)
	CapPPCOSI                   = Cap(52)
	CapPPCUnsetIRQ              = Cap(53)
	CapEnableCap                = Cap(54)
	CapXSave                    = Cap(55)
	CapXCRS                     = Cap(56)
	CapPPCGetPVInfo             = Cap(57)
	CapPPCIRQLevel              = Cap(58)
	CapAsyncPF                  = Cap(59)
	CapTSCControl               = Cap(60)
	CapGetTSCKHz                = Cap(61)
	CapPPCBookeSRegs            = Cap(62)
	CapSPAPRTCE                 = Cap(63)
	CapPPCSMT                   = Cap(64)
	CapPPCRMA                   = Cap(65)
	CapMaxVCPUs                 = Cap(66)
	CapPPCHIOR                  = Cap(67)
	CapPPCPAPR                  = Cap(68)
	CapSWTLB                    = Cap(69)
	CapOneReg                   = Cap(70)
	CapS390GMap                 = Cap(71)
	CapTSCDeadlineTimer         = Cap(72)
	CapS390UControl             = Cap(73)
	CapSyncRegs                 = Cap(74)
	CapPCI23                    = Cap(75)
	CapKVMClockCtrl             = Cap(76)
	CapSignalMSI                = Cap(77)
	CapPPCGetSMMUInfo           = Cap(78)
	CapS390COW                  = Cap(79)
	CapPPCAllocHTAB             = Cap(80)
	CapReadonlyMem              = Cap(81)
	CapIRQFDResample            = Cap(82)
	CapPPCBookeWatchdog         = Cap(83)
	CapPPCHTABFD                = Cap(84)
	CapS390CSSSupport           = Cap(85)
	CapPPCEPR                   = Cap(86)
	CapARMPSCI                  = Cap(87)
	CapARMSetDeviceAddr         = Cap(88)
	CapDeviceCtrl               = Cap(89)
	CapIRQMPIC                  = Cap(90)
	CapPPCRTAS                  = Cap(91)
	CapIRQXICS                  = Cap(92)
	CapARMEL132Bit              = Cap(93)
	CapSPAPRMultitce            = Cap(94)
	CapExtEmulCPUID             = Cap(95)
	CapHyperVTime               = Cap(96)
	CapIOAPICPolarityIgnored    = Cap(97)
	CapEnableCapVM              = Cap(98)
	CapS390IRQChip              = Cap(99)
	CapIOEventFDNoLength        = Cap(100)
	CapVMAttributes             = Cap(101)
	CapARMPSCI02                = Cap(102)
	CapPPCFixupHCall            = Cap(103)
	CapPPCEnableHCall           = Cap(104)
	CapCheckExtensionVM         = Cap(105)
	CapS390UserSIGP             = Cap(106)
	CapS390VectorRegisters      = Cap(107)
	CapS390MemOp                = Cap(108)
	CapS390UserSTSI             = Cap(109)
	CapS390SKeys                = Cap(110)
	CapMIPSFPU                  = Cap(111)
	CapMIPSMSA                  = Cap(112)
	CapS390InjectIRQ            = Cap(113)
	CapS390IRQState             = Cap(114)
	CapPPCHWRNG                 = Cap(115)
	CapDisableQuirks            = Cap(116)
	CapX86SMM                   = Cap(117)
	CapMultiAddressSpace        = Cap(118)
	CapGuestDebugHWBPS          = Cap(119)
	CapGuestDebugHWWPS          = Cap(120)
	CapSplitIRQChip             = Cap(121)
	CapIOEventFDAnyLength       = Cap(122)
	CapHyperVSYNIC              = Cap(123)
	CapS390RI                   = Cap(124)
	CapSPAPRTCE64               = Cap(125)
	CapARMPMUv3                 = Cap(126)
	CapVCPUAttributes           = Cap(127)
	CapMaxVCPUID                = Cap(128)
	CapX2APICAPI                = Cap(129)
	CapS390UserInstr0           = Cap(130)
	CapMSIDevid                 = Cap(131)
	CapPPCHTM                   = Cap(132)
	CapSPAPRResizeHPT           = Cap(133)
	CapPPCMMURadix              = Cap(134)
	CapPPCMMUHashV3             = Cap(135)
	CapImmediateExit            = Cap(136)
	CapMIPSVZ                   = Cap(137)
	CapMIPSTE                   = Cap(138)
	CapMIPS64Bit                = Cap(139)
	CapS390GS                   = Cap(140)
	CapS390AIS                  = Cap(141)
	CapSPAPRTCEVFIO             = Cap(142)
	CapX86DisableExits          = Cap(143)
	CapARMUserIRQ               = Cap(144)
	CapS390CMMAMigration        = Cap(145)
	CapPPCFWNMI                 = Cap(146)
	CapPPCSMTPossible           = Cap(147)
	CapHyperVSYNIC2             = Cap(148)
	CapHyperVVPIndex            = Cap(149)
	CapS390AISMigration         = Cap(150)
	CapPPCGetCPUChar            = Cap(151)
	CapS390BPB                  = Cap(152)
	CapGetMSRFeatures           = Cap(153)
	CapHyperVEventFD            = Cap(154)
	CapHyperVTLBFlush           = Cap(155)
	CapS390HPage1M              = Cap(156)
	CapNestedState              = Cap(157)
	CapARMINJECTSERRORESR       = Cap(158)
	CapMSRPlatformInfo          = Cap(159)
	CapPPCNestedHV              = Cap(160)
	CapHyperVSendIPI            = Cap(161)
	CapCoalescedPIO             = Cap(162)
	CapHyperVEnlightenedVMCS    = Cap(163)
	CapExceptionPayload         = Cap(164)
	CapARMVMIPASize             = Cap(165)
	CapManualDirtyLogProtect    = Cap(166)
	CapHyperVCPUID              = Cap(167)
	CapManualDirtyLogProtect2   = Cap(168)
	CapPPCIRQXIVE               = Cap(169)
	CapARMSVE                   = Cap(170)
	CapARMPtrauthAddress        = Cap(171)
	CapARMPtrauthGeneric        = Cap(172)
	CapPMUEventFilter           = Cap(173)
	CapARMIRQLineLayout2        = Cap(174)
	CapHyperVDirectTLBFlush     = Cap(175)
	CapPPCGuestDebugSStep       = Cap(176)
	CapARMNISVToUser            = Cap(177)
	CapARMInjectExtDABT         = Cap(178)
	CapS390VCPUResets           = Cap(179)
	CapS390Protected            = Cap(180)
	CapPPCSecureGuest           = Cap(181)
	CapHaltPoll                 = Cap(182)
	CapAsyncPFInt               = Cap(183)
	CapLastCPU                  = Cap(184)
	CapSmallerMaxPhyAddr        = Cap(185)
	CapS390Diag318              = Cap(186)
	CapStealTime                = Cap(187)
	CapX86UserSpaceMSR          = Cap(188)
	CapX86MSRFilter             = Cap(189)
	CapEnforcePVFeatureCPUID    = Cap(190)
	CapSysHyperVCPUID           = Cap(191)
	CapDirtyLogRing             = Cap(192)
	CapX86BusLockExit           = Cap(193)
	CapPPCDAWR1                 = Cap(194)
	CapSetGuestDebug2           = Cap(195)
	CapSGXAttribute             = Cap(196)
	CapVMCopyEncContextFrom     = Cap(197)
	CapPTPKVM                   = Cap(198)
	CapHyperVEnforceCPUID       = Cap(199)
	CapSRegs2                   = Cap(200)
	CapExitHypercall            = Cap(201)
	CapPPCRPTInvalidate         = Cap(202)
	CapBinaryStatsFD            = Cap(203)
	CapExitOnEmulationFailure   = Cap(204)
	CapARMMTE                   = Cap(205)
	CapVMMoveEncContextFrom     = Cap(206)
	CapVMGPABits                = Cap(207)
	CapXSave2                   = Cap(208)
	CapSysAttributes            = Cap(209)
	CapPPCAILMode3              = Cap(210)
	CapS390MemOpExtension       = Cap(211)
	CapPMUCapability            = Cap(212)
	CapDisableQuirks2           = Cap(213)
	CapVMTSCControl             = Cap(214)
	CapSystemEventData          = Cap(215)
	CapARMSuspend               = Cap(216)
	CapS390ProtectedDump        = Cap(217)
	CapX86TripleFaultEvent      = Cap(218)
	CapX86NotifyVMExit          = Cap(219)
	CapVMDisableNXHugePages     = Cap(220)
	CapS390ZPCIOp               = Cap(221)
	CapS390CPUTopology          = Cap(222)
	CapDirtyLogRingAcqRel       = Cap(223)
)

var allCaps = []Cap{
	CapIRQChip,
	CapHLT,
	CapMMUShadowCacheControl,
	CapUserMemory,
	CapSetTSSAddr,
	CapVAPIC,
	CapExtCPUID,
	CapClockSource,
	CapNrVCPUs,
	CapNrMemslots,
	CapPIT,
	CapNopIODelay,
	CapPVMMU,
	CapMPState,
	CapCoalescedMMIO,
	CapSyncMMU,
	CapIOMMU,
	CapDestroyMemoryRegionWorks,
	CapUserNMI,
	CapSetGuestDebug,
	CapReinjectControl,
	CapIRQRouting,
	CapIRQInjectStatus,
	CapAssignDevIRQ,
	CapJoinMemoryRegionsWorks,
	CapMCE,
	CapIRQFD,
	CapPIT2,
	CapSetBootCPUID,
	CapPITState2,
	CapIOEventFD,
	CapSetIdentityMapAddr,
	CapXenHVM,
	CapAdjustClock,
	CapInternalErrorData,
	CapVCPUEvents,
	CapS390PSW,
	CapPPCSegState,
	CapHyperV,
	CapHyperVVAPIC,
	CapHyperVSpin,
	CapPCISegment,
	CapPPCPairedSingles,
	CapIntrShadow,
	CapDebugRegs,
	CapX86RobustSingleStep,
	CapPPCOSI,
	CapPPCUnsetIRQ,
	CapEnableCap,
	CapXSave,
	CapXCRS,
	CapPPCGetPVInfo,
	CapPPCIRQLevel,
	CapAsyncPF,
	CapTSCControl,
	CapGetTSCKHz,
	CapPPCBookeSRegs,
	CapSPAPRTCE,
	CapPPCSMT,
	CapPPCRMA,
	CapMaxVCPUs,
	CapPPCHIOR,
	CapPPCPAPR,
	CapSWTLB,
	CapOneReg,
	CapS390GMap,
	CapTSCDeadlineTimer,
	CapS390UControl,
	CapSyncRegs,
	CapPCI23,
	CapKVMClockCtrl,
	CapSignalMSI,
	CapPPCGetSMMUInfo,
	CapS390COW,
	CapPPCAllocHTAB,
	CapReadonlyMem,
	CapIRQFDResample,
	CapPPCBookeWatchdog,
	CapPPCHTABFD,
	CapS390CSSSupport,
	CapPPCEPR,
	CapARMPSCI,
	CapARMSetDeviceAddr,
	CapDeviceCtrl,
	CapIRQMPIC,
	CapPPCRTAS,
	CapIRQXICS,
	CapARMEL132Bit,
	CapSPAPRMultitce,
	CapExtEmulCPUID,
	CapHyperVTime,
	CapIOAPICPolarityIgnored,
	CapEnableCapVM,
	CapS390IRQChip,
	CapIOEventFDNoLength,
	CapVMAttributes,
	CapARMPSCI02,
	CapPPCFixupHCall,
	CapPPCEnableHCall,
	CapCheckExtensionVM,
	CapS390UserSIGP,
	CapS390VectorRegisters,
	CapS390MemOp,
	CapS390UserSTSI,
	CapS390SKeys,
	CapMIPSFPU,
	CapMIPSMSA,
	CapS390InjectIRQ,
	CapS390IRQState,
	CapPPCHWRNG,
	CapDisableQuirks,
	CapX86SMM,
	CapMultiAddressSpace,
	CapGuestDebugHWBPS,
	CapGuestDebugHWWPS,
	CapSplitIRQChip,
	CapIOEventFDAnyLength,
	CapHyperVSYNIC,
	CapS390RI,
	CapSPAPRTCE64,
	CapARMPMUv3,
	CapVCPUAttributes,
	CapMaxVCPUID,
	CapX2APICAPI,
	CapS390UserInstr0,
	CapMSIDevid,
	CapPPCHTM,
	CapSPAPRResizeHPT,
	CapPPCMMURadix,
	CapPPCMMUHashV3,
	CapImmediateExit,
	CapMIPSVZ,
	CapMIPSTE,
	CapMIPS64Bit,
	CapS390GS,
	CapS390AIS,
	CapSPAPRTCEVFIO,
	CapX86DisableExits,
	CapARMUserIRQ,
	CapS390CMMAMigration,
	CapPPCFWNMI,
	CapPPCSMTPossible,
	CapHyperVSYNIC2,
	CapHyperVVPIndex,
	CapS390AISMigration,
	CapPPCGetCPUChar,
	CapS390BPB,
	CapGetMSRFeatures,
	CapHyperVEventFD,
	CapHyperVTLBFlush,
	CapS390HPage1M,
	CapNestedState,
	CapARMINJECTSERRORESR,
	CapMSRPlatformInfo,
	CapPPCNestedHV,
	CapHyperVSendIPI,
	CapCoalescedPIO,
	CapHyperVEnlightenedVMCS,
	CapExceptionPayload,
	CapARMVMIPASize,
	CapManualDirtyLogProtect,
	CapHyperVCPUID,
	CapManualDirtyLogProtect2,
	CapPPCIRQXIVE,
	CapARMSVE,
	CapARMPtrauthAddress,
	CapARMPtrauthGeneric,
	CapPMUEventFilter,
	CapARMIRQLineLayout2,
	CapHyperVDirectTLBFlush,
	CapPPCGuestDebugSStep,
	CapARMNISVToUser,
	CapARMInjectExtDABT,
	CapS390VCPUResets,
	CapS390Protected,
	CapPPCSecureGuest,
	CapHaltPoll,
	CapAsyncPFInt,
	CapLastCPU,
	CapSmallerMaxPhyAddr,
	CapS390Diag318,
	CapStealTime,
	CapX86UserSpaceMSR,
	CapX86MSRFilter,
	CapEnforcePVFeatureCPUID,
	CapSysHyperVCPUID,
	CapDirtyLogRing,
	CapX86BusLockExit,
	CapPPCDAWR1,
	CapSetGuestDebug2,
	CapSGXAttribute,
	CapVMCopyEncContextFrom,
	CapPTPKVM,
	CapHyperVEnforceCPUID,
	CapSRegs2,
	CapExitHypercall,
	CapPPCRPTInvalidate,
	CapBinaryStatsFD,
	CapExitOnEmulationFailure,
	CapARMMTE,
	CapVMMoveEncContextFrom,
	CapVMGPABits,
	CapXSave2,
	CapSysAttributes,
	CapPPCAILMode3,
	CapS390MemOpExtension,
	CapPMUCapability,
	CapDisableQuirks2,
	CapVMTSCControl,
	CapSystemEventData,
	CapARMSuspend,
	CapS390ProtectedDump,
	CapX86TripleFaultEvent,
	CapX86NotifyVMExit,
	CapVMDisableNXHugePages,
	CapS390ZPCIOp,
	CapS390CPUTopology,
	CapDirtyLogRingAcqRel,
}

// String returns the name of the capability.
func (c Cap) String() string {
	switch c {
	case CapIRQChip:
		return "KVM_CAP_IRQCHIP"
	case CapHLT:
		return "KVM_CAP_HLT"
	case CapMMUShadowCacheControl:
		return "KVM_CAP_MMU_SHADOW_CACHE_CONTROL"
	case CapUserMemory:
		return "KVM_CAP_USER_MEMORY"
	case CapSetTSSAddr:
		return "KVM_CAP_SET_TSS_ADDR"
	case CapVAPIC:
		return "KVM_CAP_VAPIC"
	case CapExtCPUID:
		return "KVM_CAP_EXT_CPUID"
	case CapClockSource:
		return "KVM_CAP_CLOCKSOURCE"
	case CapNrVCPUs:
		return "KVM_CAP_NR_VCPUS"
	case CapNrMemslots:
		return "KVM_CAP_NR_MEMSLOTS"
	case CapPIT:
		return "KVM_CAP_PIT"
	case CapNopIODelay:
		return "KVM_CAP_NOP_IO_DELAY"
	case CapPVMMU:
		return "KVM_CAP_PV_MMU"
	case CapMPState:
		return "KVM_CAP_MP_STATE"
	case CapCoalescedMMIO:
		return "KVM_CAP_COALESCED_MMIO"
	case CapSyncMMU:
		return "KVM_CAP_SYNC_MMU"
	case CapIOMMU:
		return "KVM_CAP_IOMMU"
	case CapDestroyMemoryRegionWorks:
		return "KVM_CAP_DESTROY_MEMORY_REGION_WORKS"
	case CapUserNMI:
		return "KVM_CAP_USER_NMI"
	case CapSetGuestDebug:
		return "KVM_CAP_SET_GUEST_DEBUG"
	case CapReinjectControl:
		return "KVM_CAP_REINJECT_CONTROL"
	case CapIRQRouting:
		return "KVM_CAP_IRQ_ROUTING"
	case CapIRQInjectStatus:
		return "KVM_CAP_IRQ_INJECT_STATUS"
	case CapAssignDevIRQ:
		return "KVM_CAP_ASSIGN_DEV_IRQ"
	case CapJoinMemoryRegionsWorks:
		return "KVM_CAP_JOIN_MEMORY_REGIONS_WORKS"
	case CapMCE:
		return "KVM_CAP_MCE"
	case CapIRQFD:
		return "KVM_CAP_IRQFD"
	case CapPIT2:
		return "KVM_CAP_PIT2"
	case CapSetBootCPUID:
		return "KVM_CAP_SET_BOOT_CPU_ID"
	case CapPITState2:
		return "KVM_CAP_PIT_STATE2"
	case CapIOEventFD:
		return "KVM_CAP_IOEVENTFD"
	case CapSetIdentityMapAddr:
		return "KVM_CAP_SET_IDENTITY_MAP_ADDR"
	case CapXenHVM:
		return "KVM_CAP_XEN_HVM"
	case CapAdjustClock:
		return "KVM_CAP_ADJUST_CLOCK"
	case CapInternalErrorData:
		return "KVM_CAP_INTERNAL_ERROR_DATA"
	case CapVCPUEvents:
		return "KVM_CAP_VCPU_EVENTS"
	case CapS390PSW:
		return "KVM_CAP_S390_PSW"
	case CapPPCSegState:
		return "KVM_CAP_PPC_SEGSTATE"
	case CapHyperV:
		return "KVM_CAP_HYPERV"
	case CapHyperVVAPIC:
		return "KVM_CAP_HYPERV_VAPIC"
	case CapHyperVSpin:
		return "KVM_CAP_HYPERV_SPIN"
	case CapPCISegment:
		return "KVM_CAP_PCI_SEGMENT"
	case CapPPCPairedSingles:
		return "KVM_CAP_PPC_PAIRED_SINGLES"
	case CapIntrShadow:
		return "KVM_CAP_INTR_SHADOW"
	case CapDebugRegs:
		return "KVM_CAP_DEBUGREGS"
	case CapX86RobustSingleStep:
		return "KVM_CAP_X86_ROBUST_SINGLESTEP"
	case CapPPCOSI:
		return "KVM_CAP_PPC_OSI"
	case CapPPCUnsetIRQ:
		return "KVM_CAP_PPC_UNSET_IRQ"
	case CapEnableCap:
		return "KVM_CAP_ENABLE_CAP"
	case CapXSave:
		return "KVM_CAP_XSAVE"
	case CapXCRS:
		return "KVM_CAP_XCRS"
	case CapPPCGetPVInfo:
		return "KVM_CAP_PPC_GET_PVINFO"
	case CapPPCIRQLevel:
		return "KVM_CAP_PPC_IRQ_LEVEL"
	case CapAsyncPF:
		return "KVM_CAP_ASYNC_PF"
	case CapTSCControl:
		return "KVM_CAP_TSC_CONTROL"
	case CapGetTSCKHz:
		return "KVM_CAP_GET_TSC_KHZ"
	case CapPPCBookeSRegs:
		return "KVM_CAP_PPC_BOOKE_SREGS"
	case CapSPAPRTCE:
		return "KVM_CAP_SPAPR_TCE"
	case CapPPCSMT:
		return "KVM_CAP_PPC_SMT"
	case CapPPCRMA:
		return "KVM_CAP_PPC_RMA"
	case CapMaxVCPUs:
		return "KVM_CAP_MAX_VCPUS"
	case CapPPCHIOR:
		return "KVM_CAP_PPC_HIOR"
	case CapPPCPAPR:
		return "KVM_CAP_PPC_PAPR"
	case CapSWTLB:
		return "KVM_CAP_SW_TLB"
	case CapOneReg:
		return "KVM_CAP_ONE_REG"
	case CapS390GMap:
		return "KVM_CAP_S390_GMAP"
	case CapTSCDeadlineTimer:
		return "KVM_CAP_TSC_DEADLINE_TIMER"
	case CapS390UControl:
		return "KVM_CAP_S390_UCONTROL"
	case CapSyncRegs:
		return "KVM_CAP_SYNC_REGS"
	case CapPCI23:
		return "KVM_CAP_PCI_2_3"
	case CapKVMClockCtrl:
		return "KVM_CAP_KVMCLOCK_CTRL"
	case CapSignalMSI:
		return "KVM_CAP_SIGNAL_MSI"
	case CapPPCGetSMMUInfo:
		return "KVM_CAP_PPC_GET_SMMU_INFO"
	case CapS390COW:
		return "KVM_CAP_S390_COW"
	case CapPPCAllocHTAB:
		return "KVM_CAP_PPC_ALLOC_HTAB"
	case CapReadonlyMem:
		return "KVM_CAP_READONLY_MEM"
	case CapIRQFDResample:
		return "KVM_CAP_IRQFD_RESAMPLE"
	case CapPPCBookeWatchdog:
		return "KVM_CAP_PPC_BOOKE_WATCHDOG"
	case CapPPCHTABFD:
		return "KVM_CAP_PPC_HTAB_FD"
	case CapS390CSSSupport:
		return "KVM_CAP_S390_CSS_SUPPORT"
	case CapPPCEPR:
		return "KVM_CAP_PPC_EPR"
	case CapARMPSCI:
		return "KVM_CAP_ARM_PSCI"
	case CapARMSetDeviceAddr:
		return "KVM_CAP_ARM_SET_DEVICE_ADDR"
	case CapDeviceCtrl:
		return "KVM_CAP_DEVICE_CTRL"
	case CapIRQMPIC:
		return "KVM_CAP_IRQ_MPIC"
	case CapPPCRTAS:
		return "KVM_CAP_PPC_RTAS"
	case CapIRQXICS:
		return "KVM_CAP_IRQ_XICS"
	case CapARMEL132Bit:
		return "KVM_CAP_ARM_EL1_32BIT"
	case CapSPAPRMultitce:
		return "KVM_CAP_SPAPR_MULTITCE"
	case CapExtEmulCPUID:
		return "KVM_CAP_EXT_EMUL_CPUID"
	case CapHyperVTime:
		return "KVM_CAP_HYPERV_TIME"
	case CapIOAPICPolarityIgnored:
		return "KVM_CAP_IOAPIC_POLARITY_IGNORED"
	case CapEnableCapVM:
		return "KVM_CAP_ENABLE_CAP_VM"
	case CapS390IRQChip:
		return "KVM_CAP_S390_IRQCHIP"
	case CapIOEventFDNoLength:
		return "KVM_CAP_IOEVENTFD_NO_LENGTH"
	case CapVMAttributes:
		return "KVM_CAP_VM_ATTRIBUTES"
	case CapARMPSCI02:
		return "KVM_CAP_ARM_PSCI_0_2"
	case CapPPCFixupHCall:
		return "KVM_CAP_PPC_FIXUP_HCALL"
	case CapPPCEnableHCall:
		return "KVM_CAP_PPC_ENABLE_HCALL"
	case CapCheckExtensionVM:
		return "KVM_CAP_CHECK_EXTENSION_VM"
	case CapS390UserSIGP:
		return "KVM_CAP_S390_USER_SIGP"
	case CapS390VectorRegisters:
		return "KVM_CAP_S390_VECTOR_REGISTERS"
	case CapS390MemOp:
		return "KVM_CAP_S390_MEM_OP"
	case CapS390UserSTSI:
		return "KVM_CAP_S390_USER_STSI"
	case CapS390SKeys:
		return "KVM_CAP_S390_SKEYS"
	case CapMIPSFPU:
		return "KVM_CAP_MIPS_FPU"
	case CapMIPSMSA:
		return "KVM_CAP_MIPS_MSA"
	case CapS390InjectIRQ:
		return "KVM_CAP_S390_INJECT_IRQ"
	case CapS390IRQState:
		return "KVM_CAP_S390_IRQ_STATE"
	case CapPPCHWRNG:
		return "KVM_CAP_PPC_HWRNG"
	case CapDisableQuirks:
		return "KVM_CAP_DISABLE_QUIRKS"
	case CapX86SMM:
		return "KVM_CAP_X86_SMM"
	case CapMultiAddressSpace:
		return "KVM_CAP_MULTI_ADDRESS_SPACE"
	case CapGuestDebugHWBPS:
		return "KVM_CAP_GUEST_DEBUG_HW_BPS"
	case CapGuestDebugHWWPS:
		return "KVM_CAP_GUEST_DEBUG_HW_WPS"
	case CapSplitIRQChip:
		return "KVM_CAP_SPLIT_IRQCHIP"
	case CapIOEventFDAnyLength:
		return "KVM_CAP_IOEVENTFD_ANY_LENGTH"
	case CapHyperVSYNIC:
		return "KVM_CAP_HYPERV_SYNIC"
	case CapS390RI:
		return "KVM_CAP_S390_RI"
	case CapSPAPRTCE64:
		return "KVM_CAP_SPAPR_TCE_64"
	case CapARMPMUv3:
		return "KVM_CAP_ARM_PMU_V3"
	case CapVCPUAttributes:
		return "KVM_CAP_VCPU_ATTRIBUTES"
	case CapMaxVCPUID:
		return "KVM_CAP_MAX_VCPU_ID"
	case CapX2APICAPI:
		return "KVM_CAP_X2APIC_API"
	case CapS390UserInstr0:
		return "KVM_CAP_S390_USER_INSTR0"
	case CapMSIDevid:
		return "KVM_CAP_MSI_DEVID"
	case CapPPCHTM:
		return "KVM_CAP_PPC_HTM"
	case CapSPAPRResizeHPT:
		return "KVM_CAP_SPAPR_RESIZE_HPT"
	case CapPPCMMURadix:
		return "KVM_CAP_PPC_MMU_RADIX"
	case CapPPCMMUHashV3:
		return "KVM_CAP_PPC_MMU_HASH_V3"
	case CapImmediateExit:
		return "KVM_CAP_IMMEDIATE_EXIT"
	case CapMIPSVZ:
		return "KVM_CAP_MIPS_VZ"
	case CapMIPSTE:
		return "KVM_CAP_MIPS_TE"
	case CapMIPS64Bit:
		return "KVM_CAP_MIPS_64BIT"
	case CapS390GS:
		return "KVM_CAP_S390_GS"
	case CapS390AIS:
		return "KVM_CAP_S390_AIS"
	case CapSPAPRTCEVFIO:
		return "KVM_CAP_SPAPR_TCE_VFIO"
	case CapX86DisableExits:
		return "KVM_CAP_X86_DISABLE_EXITS"
	case CapARMUserIRQ:
		return "KVM_CAP_ARM_USER_IRQ"
	case CapS390CMMAMigration:
		return "KVM_CAP_S390_CMMA_MIGRATION"
	case CapPPCFWNMI:
		return "KVM_CAP_PPC_FWNMI"
	case CapPPCSMTPossible:
		return "KVM_CAP_PPC_SMT_POSSIBLE"
	case CapHyperVSYNIC2:
		return "KVM_CAP_HYPERV_SYNIC2"
	case CapHyperVVPIndex:
		return "KVM_CAP_HYPERV_VP_INDEX"
	case CapS390AISMigration:
		return "KVM_CAP_S390_AIS_MIGRATION"
	case CapPPCGetCPUChar:
		return "KVM_CAP_PPC_GET_CPU_CHAR"
	case CapS390BPB:
		return "KVM_CAP_S390_BPB"
	case CapGetMSRFeatures:
		return "KVM_CAP_GET_MSR_FEATURES"
	case CapHyperVEventFD:
		return "KVM_CAP_HYPERV_EVENTFD"
	case CapHyperVTLBFlush:
		return "KVM_CAP_HYPERV_TLBFLUSH"
	case CapS390HPage1M:
		return "KVM_CAP_S390_HPAGE_1M"
	case CapNestedState:
		return "KVM_CAP_NESTED_STATE"
	case CapARMINJECTSERRORESR:
		return "KVM_CAP_ARM_INJECT_SERROR_ESR"
	case CapMSRPlatformInfo:
		return "KVM_CAP_MSR_PLATFORM_INFO"
	case CapPPCNestedHV:
		return "KVM_CAP_PPC_NESTED_HV"
	case CapHyperVSendIPI:
		return "KVM_CAP_HYPERV_SEND_IPI"
	case CapCoalescedPIO:
		return "KVM_CAP_COALESCED_PIO"
	case CapHyperVEnlightenedVMCS:
		return "KVM_CAP_HYPERV_ENLIGHTENED_VMCS"
	case CapExceptionPayload:
		return "KVM_CAP_EXCEPTION_PAYLOAD"
	case CapARMVMIPASize:
		return "KVM_CAP_ARM_VM_IPA_SIZE"
	case CapManualDirtyLogProtect:
		return "KVM_CAP_MANUAL_DIRTY_LOG_PROTECT"
	case CapHyperVCPUID:
		return "KVM_CAP_HYPERV_CPUID"
	case CapManualDirtyLogProtect2:
		return "KVM_CAP_MANUAL_DIRTY_LOG_PROTECT2"
	case CapPPCIRQXIVE:
		return "KVM_CAP_PPC_IRQ_XIVE"
	case CapARMSVE:
		return "KVM_CAP_ARM_SVE"
	case CapARMPtrauthAddress:
		return "KVM_CAP_ARM_PTRAUTH_ADDRESS"
	case CapARMPtrauthGeneric:
		return "KVM_CAP_ARM_PTRAUTH_GENERIC"
	case CapPMUEventFilter:
		return "KVM_CAP_PMU_EVENT_FILTER"
	case CapARMIRQLineLayout2:
		return "KVM_CAP_ARM_IRQ_LINE_LAYOUT_2"
	case CapHyperVDirectTLBFlush:
		return "KVM_CAP_HYPERV_DIRECT_TLBFLUSH"
	case CapPPCGuestDebugSStep:
		return "KVM_CAP_PPC_GUEST_DEBUG_SSTEP"
	case CapARMNISVToUser:
		return "KVM_CAP_ARM_NISV_TO_USER"
	case CapARMInjectExtDABT:
		return "KVM_CAP_ARM_INJECT_EXT_DABT"
	case CapS390VCPUResets:
		return "KVM_CAP_S390_VCPU_RESETS"
	case CapS390Protected:
		return "KVM_CAP_S390_PROTECTED"
	case CapPPCSecureGuest:
		return "KVM_CAP_PPC_SECURE_GUEST"
	case CapHaltPoll:
		return "KVM_CAP_HALT_POLL"
	case CapAsyncPFInt:
		return "KVM_CAP_ASYNC_PF_INT"
	case CapLastCPU:
		return "KVM_CAP_LAST_CPU"
	case CapSmallerMaxPhyAddr:
		return "KVM_CAP_SMALLER_MAXPHYADDR"
	case CapS390Diag318:
		return "KVM_CAP_S390_DIAG318"
	case CapStealTime:
		return "KVM_CAP_STEAL_TIME"
	case CapX86UserSpaceMSR:
		return "KVM_CAP_X86_USER_SPACE_MSR"
	case CapX86MSRFilter:
		return "KVM_CAP_X86_MSR_FILTER"
	case CapEnforcePVFeatureCPUID:
		return "KVM_CAP_ENFORCE_PV_FEATURE_CPUID"
	case CapSysHyperVCPUID:
		return "KVM_CAP_SYS_HYPERV_CPUID"
	case CapDirtyLogRing:
		return "KVM_CAP_DIRTY_LOG_RING"
	case CapX86BusLockExit:
		return "KVM_CAP_X86_BUS_LOCK_EXIT"
	case CapPPCDAWR1:
		return "KVM_CAP_PPC_DAWR1"
	case CapSetGuestDebug2:
		return "KVM_CAP_SET_GUEST_DEBUG2"
	case CapSGXAttribute:
		return "KVM_CAP_SGX_ATTRIBUTE"
	case CapVMCopyEncContextFrom:
		return "KVM_CAP_VM_COPY_ENC_CONTEXT_FROM"
	case CapPTPKVM:
		return "KVM_CAP_PTP_KVM"
	case CapHyperVEnforceCPUID:
		return "KVM_CAP_HYPERV_ENFORCE_CPUID"
	case CapSRegs2:
		return "KVM_CAP_SREGS2"
	case CapExitHypercall:
		return "KVM_CAP_EXIT_HYPERCALL"
	case CapPPCRPTInvalidate:
		return "KVM_CAP_PPC_RPT_INVALIDATE"
	case CapBinaryStatsFD:
		return "KVM_CAP_BINARY_STATS_FD"
	case CapExitOnEmulationFailure:
		return "KVM_CAP_EXIT_ON_EMULATION_FAILURE"
	case CapARMMTE:
		return "KVM_CAP_ARM_MTE"
	case CapVMMoveEncContextFrom:
		return "KVM_CAP_VM_MOVE_ENC_CONTEXT_FROM"
	case CapVMGPABits:
		return "KVM_CAP_VM_GPA_BITS"
	case CapXSave2:
		return "KVM_CAP_XSAVE2"
	case CapSysAttributes:
		return "KVM_CAP_SYS_ATTRIBUTES"
	case CapPPCAILMode3:
		return "KVM_CAP_PPC_AIL_MODE_3"
	case CapS390MemOpExtension:
		return "KVM_CAP_S390_MEM_OP_EXTENSION"
	case CapPMUCapability:
		return "KVM_CAP_PMU_CAPABILITY"
	case CapDisableQuirks2:
		return "KVM_CAP_DISABLE_QUIRKS2"
	case CapVMTSCControl:
		return "KVM_CAP_VM_TSC_CONTROL"
	case CapSystemEventData:
		return "KVM_CAP_SYSTEM_EVENT_DATA"
	case CapARMSuspend:
		return "KVM_CAP_ARM_SYSTEM_SUSPEND"
	case CapS390ProtectedDump:
		return "KVM_CAP_S390_PROTECTED_DUMP"
	case CapX86TripleFaultEvent:
		return "KVM_CAP_X86_TRIPLE_FAULT_EVENT"
	case CapX86NotifyVMExit:
		return "KVM_CAP_X86_NOTIFY_VMEXIT"
	case CapVMDisableNXHugePages:
		return "KVM_CAP_VM_DISABLE_NX_HUGE_PAGES"
	case CapS390ZPCIOp:
		return "KVM_CAP_S390_ZPCI_OP"
	case CapS390CPUTopology:
		return "KVM_CAP_S390_CPU_TOPOLOGY"
	case CapDirtyLogRingAcqRel:
		return "KVM_CAP_DIRTY_LOG_RING_ACQ_REL"
	default:
		return fmt.Sprintf("Cap(%d)", c)
	}
}

const (
	ExitUnknown       = Exit(0)
	ExitException     = Exit(1)
	ExitIO            = Exit(2)
	ExitHypercall     = Exit(3)
	ExitDebug         = Exit(4)
	ExitHLT           = Exit(5)
	ExitMMIO          = Exit(6)
	ExitIRQWindowOpen = Exit(7)
	ExitShutdown      = Exit(8)
	ExitFailEntry     = Exit(9)
	ExitIntr          = Exit(10)
	ExitSetTPR        = Exit(11)
	ExitTPRAccess     = Exit(12)
	ExitS390SIEIC     = Exit(13)
	ExitS390Reset     = Exit(14)
	ExitDCR           = Exit(15)
	ExitNMI           = Exit(16)
	ExitInternalError = Exit(17)
	ExitOSI           = Exit(18)
	ExitPAPRHCALL     = Exit(19)
	ExitS390UControl  = Exit(20)
	ExitWatchdog      = Exit(21)
	ExitS390TSCH      = Exit(22)
	ExitEPR           = Exit(23)
	ExitSystemEvent   = Exit(24)
	ExitS390STSI      = Exit(25)
	ExitIOAPICEOI     = Exit(26)
	ExitHyperV        = Exit(27)
	ExitARMNISV       = Exit(28)
	ExitX86RDMSR      = Exit(29)
	ExitX86WRMSR      = Exit(30)
	ExitDirtyRingFull = Exit(31)
	ExitAPResetHold   = Exit(32)
	ExitX86BusLock    = Exit(33)
	ExitXen           = Exit(34)
	ExitRISCVSBI      = Exit(35)
	ExitRISCVCSR      = Exit(36)
	ExitNotify        = Exit(37)
)

// String returns the name of the exit reason.
func (e Exit) String() string {
	switch e {
	case ExitUnknown:
		return "KVM_EXIT_UNKNOWN"
	case ExitException:
		return "KVM_EXIT_EXCEPTION"
	case ExitIO:
		return "KVM_EXIT_IO"
	case ExitHypercall:
		return "KVM_EXIT_HYPERCALL"
	case ExitDebug:
		return "KVM_EXIT_DEBUG"
	case ExitHLT:
		return "KVM_EXIT_HLT"
	case ExitMMIO:
		return "KVM_EXIT_MMIO"
	case ExitIRQWindowOpen:
		return "KVM_EXIT_IRQ_WINDOW_OPEN"
	case ExitShutdown:
		return "KVM_EXIT_SHUTDOWN"
	case ExitFailEntry:
		return "KVM_EXIT_FAIL_ENTRY"
	case ExitIntr:
		return "KVM_EXIT_INTR"
	case ExitSetTPR:
		return "KVM_EXIT_SET_TPR"
	case ExitTPRAccess:
		return "KVM_EXIT_TPR_ACCESS"
	case ExitS390SIEIC:
		return "KVM_EXIT_S390_SIEIC"
	case ExitS390Reset:
		return "KVM_EXIT_S390_RESET"
	case ExitDCR:
		return "KVM_EXIT_DCR"
	case ExitNMI:
		return "KVM_EXIT_NMI"
	case ExitInternalError:
		return "KVM_EXIT_INTERNAL_ERROR"
	case ExitOSI:
		return "KVM_EXIT_OSI"
	case ExitPAPRHCALL:
		return "KVM_EXIT_PAPR_HCALL"
	case ExitS390UControl:
		return "KVM_EXIT_S390_UCONTROL"
	case ExitWatchdog:
		return "KVM_EXIT_WATCHDOG"
	case ExitS390TSCH:
		return "KVM_EXIT_S390_TSCH"
	case ExitEPR:
		return "KVM_EXIT_EPR"
	case ExitSystemEvent:
		return "KVM_EXIT_SYSTEM_EVENT"
	case ExitS390STSI:
		return "KVM_EXIT_S390_STSI"
	case ExitIOAPICEOI:
		return "KVM_EXIT_IOAPIC_EOI"
	case ExitHyperV:
		return "KVM_EXIT_HYPERV"
	case ExitARMNISV:
		return "KVM_EXIT_ARM_NISV"
	case ExitX86RDMSR:
		return "KVM_EXIT_X86_RDMSR"
	case ExitX86WRMSR:
		return "KVM_EXIT_X86_WRMSR"
	case ExitDirtyRingFull:
		return "KVM_EXIT_DIRTY_RING_FULL"
	case ExitAPResetHold:
		return "KVM_EXIT_AP_RESET_HOLD"
	case ExitX86BusLock:
		return "KVM_EXIT_X86_BUS_LOCK"
	case ExitXen:
		return "KVM_EXIT_XEN"
	case ExitRISCVSBI:
		return "KVM_EXIT_RISCV_SBI"
	case ExitRISCVCSR:
		return "KVM_EXIT_RISCV_CSR"
	case ExitNotify:
		return "KVM_EXIT_NOTIFY"
	default:
		return fmt.Sprintf("Exit(%d)", e)
	}
}

const (
	MemLogDirtyPages = 1
	MemReadonly      = 2
)

const (
	CPUIDFlagSignificantIndex = 1
	CPUIDFlagStatefulFunc     = 2
	CPUIDFlagStateReadNext    = 4
)

const (
	PITSpeakerDummy = 1
)

const (
	kGetAPIVersion          = 0xae00
	kCreateVM               = 0xae01
	kGetMSRIndexList        = 0xc004ae02
	kGetMSRFeatureIndexList = 0xc004ae0a
	kCheckExtension         = 0xae03
	kGetVCPUMmapSize        = 0xae04
	kCreateVCPU             = 0xae41
	kRun                    = 0xae80
	kGetRegs                = 0x8090ae81
	kSetRegs                = 0x4090ae82
	kGetSregs               = 0x8138ae83
	kSetSregs               = 0x4138ae84
	kGetMSRs                = 0xc008ae88
	kSetMSRs                = 0x4008ae89
	kGetFPU                 = 0x81a0ae8c
	kSetFPU                 = 0x41a0ae8d
	kCreateIRQChip          = 0xae60
	kCreatePIT2             = 0x4040ae77
	kGetClock               = 0x8030ae7c
	kSetClock               = 0x4030ae7b
	kSetUserMemoryRegion    = 0x4020ae46
	kSetTSSAddr             = 0xae47
	kSetIdentityMapAddr     = 0x4008ae48
	kGetSupportedCPUID      = 0xc008ae05
	kSetCPUID2              = 0x4008ae90
)

const (
	nrInterrupts = 256
)
