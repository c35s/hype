package mmio

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"
	"unsafe"

	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/virtio/virtq"
	"golang.org/x/sys/unix"
)

type Bus struct {
	handlers []virtio.DeviceHandler
	memAt    func(addr uint64, size int) ([]byte, error)
	notify   func(irq int) error
	devices  []*device
}

type device struct {
	bus  *Bus
	info DeviceInfo

	mu      sync.Mutex
	handler virtio.DeviceHandler
	state   deviceState

	qC [16]chan struct{}
}

type deviceState struct {
	status  uint32
	version uint32

	deviceFeaturesSel uint32
	driverFeaturesSel uint32
	driverFeatures    uint64

	queueSel uint32
	queue    [16]queueState

	intStatus uint32
}

type queueState struct {
	Ready      uint32
	NumDesc    uint32
	DescAddr   uint64 // address of the descriptor area
	DriverAddr uint64 // address of the driver area
	DeviceAddr uint64 // address of the device area
}

const (
	statusAcknowledge = 1   // recognized by the guest
	statusDriver      = 2   // the guest has a driver
	statusFeaturesOK  = 8   // features negotiated
	statusDriverOK    = 4   // ready to drive
	statusNeedsReset  = 64  // fatal device error
	statusFailed      = 128 // fatal driver error

	negotiatingFeatures = statusAcknowledge | statusDriver
	configuringQueues   = negotiatingFeatures | statusFeaturesOK
	operatingNormally   = configuringQueues | statusDriverOK
)

var le = binary.LittleEndian

// NewBus creates a new bus and installs a device for for each of the given handlers.
// The memAt callback is called when a device needs to access a virtqueue in guest memory.
// The notify callback is called when a device needs to notify the guest of a config or buffer event.
//
// Devices are assigned an IRQ and a 4K memory region. See the Devices method.
func NewBus(handlers []virtio.DeviceHandler, memAt func(addr uint64, size int) ([]byte, error), notify func(irq int) error) *Bus {
	const sz = 0x1000

	var (
		irq  = 5
		addr = uint64(0xd0000000)
	)

	b := &Bus{
		handlers: handlers,
		memAt:    memAt,
		notify:   notify,
		devices:  make([]*device, len(handlers)),
	}

	for i, h := range handlers {
		d := &device{
			bus: b,

			info: DeviceInfo{
				Type: h.GetType(),
				IRQ:  irq,
				Addr: addr,
				Size: sz,
			},

			handler: h,
		}

		for i := range d.qC {
			d.qC[i] = make(chan struct{}, 1)
		}

		b.devices[i] = d

		irq++
		addr += sz
	}

	return b
}

// HandleMMIO routes an MMIO event to the appropriate device.
// It returns (found=false, err=nil) if no device is found.
func (b *Bus) HandleMMIO(addr uint64, data []byte, isWrite bool) (found bool, err error) {
	var dev *device
	for _, d := range b.devices {
		if addr >= d.info.Addr && addr < d.info.Addr+d.info.Size {
			dev = d
			break
		}
	}

	if dev == nil {
		return false, nil
	}

	off := int(addr - dev.info.Addr)
	return true, dev.HandleMMIO(off, data, isWrite)
}

// Devices returns a slice describing the installed devices.
func (b *Bus) Devices() []DeviceInfo {
	dd := make([]DeviceInfo, len(b.devices))
	for i, d := range b.devices {
		dd[i] = d.info
	}

	return dd
}

func (d *device) HandleMMIO(off int, data []byte, isWrite bool) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	defer func() {
		if err != nil && !(d.needsReset() || d.driverFailed()) {
			notify := d.isOperatingNormally()
			d.state.status |= statusNeedsReset
			d.state.version++

			if notify {
				d.state.intStatus |= intStatusConfigChange
				if err := d.bus.notify(d.info.IRQ); err != nil {
					slog.Error("virtio config change notification failed",
						"irq", d.info.IRQ, "err", err)
				}
			}
		}
	}()

	if isWrite {
		return d.writeMMIO(off, data)
	}

	return d.readMMIO(off, data)
}

func (d *device) readMMIO(off int, p []byte) error {
	switch off {
	case regMagicValue:
		le.PutUint32(p, virtio.MagicValue)

	case regVersion:
		le.PutUint32(p, virtio.Version)

	case regDeviceID:
		le.PutUint32(p, uint32(d.handler.GetType()))

	// FIX: ?
	case regVendorID:
		le.PutUint32(p, 0xffff)

	case regDeviceFeatures:
		le.PutUint32(p, uint32(d.getFeatures()>>(32*d.state.deviceFeaturesSel)))

	case regQueueNumMax:
		le.PutUint32(p, 1<<15)

	case regQueueReady:
		le.PutUint32(p, d.selectedQueue().Ready)

	case regInterruptStatus:
		le.PutUint32(p, d.state.intStatus)

	case regStatus:
		le.PutUint32(p, d.state.status)

	case regConfigGeneration:
		le.PutUint32(p, d.state.version)

	default:
		switch {
		case off >= regDeviceConfigStart:
			if err := d.handler.ReadConfig(p, off-regDeviceConfigStart); err != nil {
				panic(err)
			}

		default:
			panic(off)
		}
	}

	return nil
}

func (d *device) writeMMIO(off int, p []byte) error {
	// if the device or driver has failed, only allow status register writes (to reset)
	if d.state.status&(statusNeedsReset|statusFailed) > 0 && off != regStatus {
		return unix.EPERM
	}

	switch off {
	case regDeviceFeaturesSel:
		return d.writeDeviceFeaturesSel(le.Uint32(p))

	case regDriverFeatures:
		return d.writeDriverFeatures(le.Uint32(p))

	case regDriverFeaturesSel:
		return d.writeDriverFeaturesSel(le.Uint32(p))

	case regQueueSel:
		return d.writeQueueSel(le.Uint32(p))

	case regQueueNum:
		return d.writeQueueNum(le.Uint32(p))

	case regQueueReady:
		return d.writeQueueReady(le.Uint32(p))

	case regQueueNotify:
		return d.writeQueueNotify(le.Uint32(p))

	case regInterruptAck:
		return d.writeInterruptAck(le.Uint32(p))

	case regStatus:
		return d.writeStatus(le.Uint32(p))

	case regQueueDescLow:
		return d.writeQueueDescLow(le.Uint32(p))

	case regQueueDescHigh:
		return d.writeQueueDescHigh(le.Uint32(p))

	case regQueueDriverLow:
		return d.writeQueueDriverLow(le.Uint32(p))

	case regQueueDriverHigh:
		return d.writeQueueDriverHigh(le.Uint32(p))

	case regQueueDeviceLow:
		return d.writeQueueDeviceLow(le.Uint32(p))

	case regQueueDeviceHigh:
		return d.writeQueueDeviceHigh(le.Uint32(p))

	default:
		panic(off)
	}
}

func (d *device) writeStatus(v uint32) error {
	if v == 0 {
		// reset
		d.state = deviceState{}
		return nil
	}

	if v&statusNeedsReset > 0 || v < d.state.status {
		panic("bad status")
	}

	d.state.status = v
	d.state.version++

	if v&statusFailed > 0 {
		panic("driver failed")
	}

	if d.isOperatingNormally() {
		if d.state.driverFeatures&virtio.RequiredFeatures != virtio.RequiredFeatures {
			panic("missing required feature bits")
		}

		if err := d.handler.Ready(d.state.driverFeatures); err != nil {
			panic(err)
		}
	}

	return nil
}

func (d *device) writeDeviceFeaturesSel(v uint32) error {
	if !d.isNegotiatingFeatures() {
		return unix.EPERM
	}

	if v > 1 {
		return unix.EINVAL
	}

	d.state.deviceFeaturesSel = v

	return nil
}

func (d *device) writeDriverFeaturesSel(v uint32) error {
	if !d.isNegotiatingFeatures() {
		return unix.EPERM
	}

	if v > 1 {
		return unix.EINVAL
	}

	d.state.driverFeaturesSel = v
	return nil
}

func (d *device) writeDriverFeatures(v uint32) error {
	if !d.isNegotiatingFeatures() {
		return unix.EPERM
	}

	d.state.driverFeatures |= uint64(v) << (32 * d.state.driverFeaturesSel)

	if d.state.driverFeatures > d.getFeatures() {
		return unix.EINVAL
	}

	return nil
}

func (d *device) writeQueueSel(v uint32) error {
	if !d.isConfiguringQueues() {
		return unix.EPERM
	}

	d.state.queueSel = v
	return nil
}

func (d *device) writeQueueNum(v uint32) error {
	if !d.isConfiguringQueues() {
		return unix.EPERM
	}

	d.selectedQueue().NumDesc = v
	return nil
}

func (d *device) writeQueueDescLow(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DescAddr |= uint64(v)
	return nil
}

func (d *device) writeQueueDescHigh(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DescAddr |= uint64(v) << 32
	return nil
}

func (d *device) writeQueueDriverLow(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DriverAddr |= uint64(v)
	return nil
}

func (d *device) writeQueueDriverHigh(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DriverAddr |= uint64(v) << 32
	return nil
}

func (d *device) writeQueueDeviceLow(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DeviceAddr |= uint64(v)
	return nil
}

func (d *device) writeQueueDeviceHigh(v uint32) error {
	if !d.isConfiguringQueues() || d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().DeviceAddr |= uint64(v) << 32
	return nil
}

func (d *device) writeQueueReady(v uint32) error {
	if !d.isConfiguringQueues() {
		return unix.EPERM
	}

	if v != 1 {
		return unix.EINVAL
	}

	if d.selectedQueue().Ready == 1 {
		return unix.EPERM
	}

	d.selectedQueue().Ready = 1
	d.state.version++

	qs := d.selectedQueue()

	rngA, err := d.bus.memAt(qs.DescAddr, int(16*qs.NumDesc))
	if err != nil {
		return err
	}

	drvA, err := d.bus.memAt(qs.DriverAddr, 4)
	if err != nil {
		return err
	}

	devA, err := d.bus.memAt(qs.DeviceAddr, 4)
	if err != nil {
		return err
	}

	var (
		ring = unsafe.Slice((*virtq.Desc)(unsafe.Pointer(&rngA[0])), qs.NumDesc)
		drvE = (*virtq.EventSuppress)(unsafe.Pointer(&drvA[0]))
		devE = (*virtq.EventSuppress)(unsafe.Pointer(&devA[0]))
	)

	vq := virtq.New(ring, drvE, devE, virtq.Config{
		MemAt: d.bus.memAt,
		Notify: func() error {
			d.mu.Lock()
			defer d.mu.Unlock()

			d.state.intStatus |= intStatusUsedBuffer
			if err := d.bus.notify(d.info.IRQ); err != nil {
				return err
			}

			return nil
		},
	})

	qn := d.state.queueSel

	go func() {
		for range d.qC[qn] {
			if err := d.handler.Handle(int(qn), vq); err != nil {
				panic(fmt.Errorf("%v: handle queue %d: %w", d.info.Type, qn, err))
			}
		}
	}()

	return nil
}

func (d *device) writeQueueNotify(v uint32) error {
	if !d.isOperatingNormally() {
		return unix.EPERM
	}

	if d.state.queue[v].Ready != 1 {
		return unix.EPERM
	}

	select {
	case d.qC[v] <- struct{}{}:
	default:
	}

	return nil
}

func (d *device) writeInterruptAck(v uint32) error {
	if !d.isOperatingNormally() {
		return unix.EPERM
	}

	// clear flags
	d.state.intStatus &^= v

	return nil
}

func (d *device) getFeatures() uint64 {
	return virtio.RequiredFeatures | d.handler.GetFeatures()
}

func (d *device) isNegotiatingFeatures() bool {
	return d.state.status == negotiatingFeatures
}

func (d *device) isConfiguringQueues() bool {
	return d.state.status == configuringQueues
}

func (d *device) isOperatingNormally() bool {
	return d.state.status == operatingNormally
}

func (d *device) needsReset() bool {
	return d.state.status&statusNeedsReset != 0
}

func (d *device) driverFailed() bool {
	return d.state.status&statusFailed != 0
}

func (d *device) selectedQueue() *queueState {
	return &d.state.queue[d.state.queueSel]
}
