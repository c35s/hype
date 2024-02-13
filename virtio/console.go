package virtio

import (
	"io"

	"github.com/c35s/hype/virtio/virtq"
)

// ConsoleDevice configures a virtio console device.
type ConsoleDevice struct {
	In  io.Reader
	Out io.Writer
}

type consoleHandler struct {
	cfg ConsoleDevice
}

const (
	consoleRxQ = 0
	consoleTxQ = 1
)

func (cfg ConsoleDevice) NewHandler() (DeviceHandler, error) {
	return &consoleHandler{cfg}, nil
}

func (h *consoleHandler) GetType() DeviceID {
	return ConsoleDeviceID
}

func (*consoleHandler) GetFeatures() uint64 {
	return 0
}

func (*consoleHandler) Ready(negotiatedFeatures uint64) error {
	return nil
}

func (h *consoleHandler) Handle(queueNum int, q *virtq.Queue) error {
	switch queueNum {
	case consoleRxQ:
		if h.cfg.In != nil {
			return h.handleRx(q)
		}

	case consoleTxQ:
		if h.cfg.Out != nil {
			return h.handleTx(q)
		}
	}

	return nil
}

func (h *consoleHandler) ReadConfig(p []byte, off int) error {
	return nil
}

func (h *consoleHandler) handleRx(q *virtq.Queue) error {
	for {
		c, err := q.Next()
		if err != nil {
			return err
		}

		if c == nil {
			break
		}

		var n int
		for i, d := range c.Desc {
			if !d.IsWO() {
				continue
			}

			buf, gbe := c.Buf(i)
			if gbe != nil {
				return gbe
			}

			n, err = h.cfg.In.Read(buf)
			break
		}

		if err != nil && err != io.EOF {
			return err
		}

		if err := c.Release(n); err != nil {
			return err
		}
	}

	return nil
}

func (h *consoleHandler) handleTx(q *virtq.Queue) error {
	for {
		c, err := q.Next()
		if err != nil {
			return err
		}

		if c == nil {
			break
		}

		for i, d := range c.Desc {
			if d.IsWO() {
				break
			}

			buf, err := c.Buf(i)
			if err != nil {
				return err
			}

			if _, err := h.cfg.Out.Write(buf); err != nil {
				return err
			}
		}

		if err := c.Release(0); err != nil {
			return err
		}
	}

	return nil
}
