package virtio

import (
	"io"

	"github.com/c35s/hype/virtio/virtq"
)

type Console struct {
	In  io.Reader
	Out io.Writer
}

const (
	consoleRxQ = 0
	consoleTxQ = 1
)

func (c *Console) GetType() DeviceID {
	return ConsoleDeviceID
}

func (*Console) GetFeatures() uint64 {
	return 0
}

func (*Console) Ready(negotiatedFeatures uint64) error {
	return nil
}

func (c *Console) Handle(queueNum int, q *virtq.Queue) error {
	switch queueNum {
	case consoleRxQ:
		if c.In != nil {
			return c.handleRx(q)
		}

	case consoleTxQ:
		if c.Out != nil {
			return c.handleTx(q)
		}
	}

	return nil
}

func (dev *Console) ReadConfig(p []byte, off int) error {
	return nil
}

func (dev *Console) handleRx(q *virtq.Queue) error {
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

			n, err = dev.In.Read(buf)
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

func (dev *Console) handleTx(q *virtq.Queue) error {
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

			if _, err := dev.Out.Write(buf); err != nil {
				return err
			}
		}

		if err := c.Release(0); err != nil {
			return err
		}
	}

	return nil
}
