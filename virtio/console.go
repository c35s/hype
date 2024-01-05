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

func (dev *Console) handleRx(q *virtq.Queue) error {
	for {
		c := q.Next()
		if c == nil {
			break
		}

		var total int
		for i := 0; i < c.Len(); i++ {
			if !c.IsWO(i) {
				panic("descriptor is not write-only")
			}

			n, err := dev.In.Read(c.Data(i))
			if err != nil {
				return err
			}

			total += n
		}

		c.Release(total)
	}

	return nil
}

func (dev *Console) handleTx(q *virtq.Queue) error {
	for {
		c := q.Next()
		if c == nil {
			break
		}

		for i := 0; i < c.Len(); i++ {
			if !c.IsRO(i) {
				panic("descriptor is not read-only")
			}

			if _, err := dev.Out.Write(c.Data(i)); err != nil {
				return err
			}
		}

		c.Release(0)
	}

	return nil
}
