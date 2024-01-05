// Package mmio implements a virtio-mmio device bus.
package mmio

import "github.com/c35s/hype/virtio"

// DeviceInfo describes an installed virtio-mmio device.
type DeviceInfo struct {
	Type virtio.DeviceID
	IRQ  int
	Addr uint64
	Size uint64
}

// interrupt status bits

const (
	intStatusUsedBuffer   = 1 << 0 // the device has used at least 1 buffer
	intStatusConfigChange = 1 << 1 // the configuration of the device has changed
)

// mmio register offsets

const (
	regMagicValue        = 0x000 // always 0x74726976 (R; "virt")
	regVersion           = 0x004 // always 0x2 (R)
	regDeviceID          = 0x008 // virtio subsystem device id (R)
	regVendorID          = 0x00c // virtio subsystem vendor id (R)
	regDeviceFeatures    = 0x010 // flags, depends on regDeviceFeaturesSel (R)
	regDeviceFeaturesSel = 0x014 // word selection for regDeviceFeatures (W)
	regDriverFeatures    = 0x020 // feature flags activated by the driver (W)
	regDriverFeaturesSel = 0x024 // word selection for regDriverFeatures (W)
	regQueueSel          = 0x030 // virtual queue index (W)
	regQueueNumMax       = 0x034 // maximum virtual queue size (R)
	regQueueNum          = 0x038 // virtual queue size (W)
	regQueueReady        = 0x044 // virtual queue ready bit (RW)
	regQueueNotify       = 0x050 // queue notifier (W)
	regInterruptStatus   = 0x060 // interrupt status (R)
	regInterruptAck      = 0x064 // interrupt acknowledge (W)
	regStatus            = 0x070 // device status (RW)
	regQueueDescLow      = 0x080 // descriptor area GPA, low word (W)
	regQueueDescHigh     = 0x084 // descriptor area GPA, high word (W)
	regQueueDriverLow    = 0x090 // driver area GPA, low word (W)
	regQueueDriverHigh   = 0x094 // driver area GPA, high word (W)
	regQueueDeviceLow    = 0x0a0 // device area GPA, low word (W)
	regQueueDeviceHigh   = 0x0a4 // device area GPA, high word (W)
	regConfigGeneration  = 0x0fc // configuration atomicity value (R)
	regDeviceConfigStart = 0x100 // device specific configuration space >= 0x100 (RW)
)
