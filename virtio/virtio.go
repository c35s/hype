package virtio

import (
	"fmt"

	"github.com/c35s/hype/virtio/virtq"
)

type DeviceConfig interface {
	NewHandler() (DeviceHandler, error)
}

type DeviceHandler interface {

	// GetType identifies the type of the device.
	GetType() DeviceID

	// GetFeatures returns additional feature bits supported by the device.
	GetFeatures() uint64

	// Ready is called after feature negotiation is complete.
	Ready(negotiatedFeatures uint64) error

	// Handle is called when new buffers are available to the device. It is
	// called in a separate goroutine per queueNum, and calls with the same
	// queueNum are do not overlap. It's fine to block in Handle. Notifications
	// are coalesced, so Handle may only be called once in response to multiple
	// driver notifications.
	Handle(queueNum int, q *virtq.Queue) error

	// ReadConfig reads the device configuration register at off into p.
	ReadConfig(p []byte, off int) error
}

// DeviceID identifies the type of a virtio device.
type DeviceID uint32

const (
	InvalidDeviceID = DeviceID(0)
	NetworkDeviceID = DeviceID(1)
	BlockDeviceID   = DeviceID(2)
	ConsoleDeviceID = DeviceID(3)
	SocketDeviceID  = DeviceID(19)
)

const (
	MagicValue = 0x74726976 // "virt"
	Version    = 0x2
)

const (

	// FIndirectDesc (VIRTIO_F_INDIRECT_DESC) "indicates that the driver can use
	// descriptors with the VIRTQ_DESC_F_INDIRECT flag set, as described in 2.6.5.3
	// Indirect Descriptors and 2.7.7 Indirect Flag: Scatter-Gather Support."
	FIndirectDesc = 1 << 28

	// FEventIdx (VIRTIO_F_EVENT_IDX) "enables the used_event and the avail_event fields
	// as described in 2.6.7, 2.6.8 and 2.7.10."
	FEventIdx = 1 << 29

	// FVersion1 (VIRTIO_F_VERSION_1) "indicates compliance with [the virtio]
	// specification, giving a simple way to detect legacy devices or drivers."
	FVersion1 = 1 << 32

	// FAccessPlatform (VIRTIO_F_ACCESS_PLATFORM) "indicates that the device can be used
	// on a platform where device access to data in memory is limited and/or translated.
	// E.g. this is the case if the device can be located behind an IOMMU that translates
	// bus addresses from the device into physical addresses in memory, if the device can
	// be limited to only access certain memory addresses or if special commands such as a
	// cache flush can be needed to synchronise data in memory with the device. Whether
	// accesses are actually limited or translated is described by platform-specific
	// means.
	//
	// If this feature bit is set to 0, then the device has same access to memory
	// addresses supplied to it as the driver has. In particular, the device will always
	// use physical addresses matching addresses used by the driver (typically meaning
	// physical addresses used by the CPU) and not translated further, and can access any
	// address supplied to it by the driver. When clear, this overrides any platform-
	// specific description of whether device access is limited or translated in any way,
	// e.g. whether an IOMMU may be present."
	FAccessPlatform = 1 << 33

	// FRingPacked (VIRTIO_F_RING_PACKED) "indicates support for the packed virtqueue
	// layout as described in 2.7 Packed Virtqueues."
	FRingPacked = 1 << 34

	// FInOrder (VIRTIO_F_IN_ORDER) "indicates that all buffers are used by the device in
	// the same order in which they have been made available."
	FInOrder = 1 << 35

	// FOrderPlatform (VIRTIO_F_ORDER_PLATFORM) "indicates that memory accesses by the
	// driver and the device are ordered in a way described by the platform. If this
	// feature bit is negotiated, the ordering in effect for any memory accesses by the
	// driver that need to be ordered in a specific way with respect to accesses by the
	// device is the one suitable for devices described by the platform. This implies that
	// the driver needs to use memory barriers suitable for devices described by the
	// platform; e.g. for the PCI transport in the case of hardware PCI devices.
	//
	// If this feature bit is not negotiated, then the device and driver are assumed to be
	// implemented in software, that is they can be assumed to run on identical CPUs in an
	// SMP configuration. Thus a weaker form of memory barriers is sufficient to yield
	// better performance."
	FOrderPlatform = 1 << 36

	// FSRIOV (VIRTIO_F_SR_IOV) "indicates that the device supports Single Root I/O
	// Virtualization. Currently only PCI devices support this feature."
	FSRIOV = 1 << 37

	// FNotificationData (VIRTIO_F_NOTIFICATION_DATA) "indicates that the driver passes
	// extra data (besides identifying the virtqueue) in its device notifications. See
	// 2.7.23 Driver notifications."
	FNotificationData = 1 << 38

	// FNotifConfigData (VIRTIO_F_NOTIF_CONFIG_DATA) "indicates that the driver uses the
	// data provided by the device as a virtqueue identifier in available buffer
	// notifications. As mentioned in section 2.9, when the driver is required to send an
	// available buffer notification to the device, it sends the virtqueue number to be
	// notified. The method of delivering notifications is transport specific. With the
	// PCI transport, the device can optionally provide a per-virtqueue value for the
	// driver to use in driver notifications, instead of the virtqueue number. Some
	// devices may benefit from this flexibility by providing, for example, an internal
	// virtqueue identifier, or an internal offset related to the virtqueue number.
	//
	// This feature indicates the availability of such value. The definition of the data
	// to be provided in driver notification and the delivery method is transport
	// specific. For more details about driver notifications over PCI see 4.1.5.2."
	FNotifConfigData = 1 << 39

	// FRingReset (VIRTIO_F_RING_RESET) "indicates that the driver can reset a queue
	// individually. See 2.6.1."
	FRingReset = 1 << 40
)

// RequiredFeatures are the feature bits negotiated for all virtio devices.
const RequiredFeatures = FVersion1 | FRingPacked | FIndirectDesc | FEventIdx

func (id DeviceID) String() string {
	switch id {
	case InvalidDeviceID:
		return "invalid"

	case NetworkDeviceID:
		return "network"

	case BlockDeviceID:
		return "block"

	case ConsoleDeviceID:
		return "console"

	case SocketDeviceID:
		return "socket"

	default:
		return fmt.Sprintf("DeviceID(%d)", id)
	}
}
