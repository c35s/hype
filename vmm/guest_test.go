package vmm_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/c35s/hype/virtio"
	"github.com/c35s/hype/vmm"
)

func TestConsole(t *testing.T) {
	GuestTest{
		Host: func(t *testing.T, runGuest func(vmm.Config)) {
			out := new(bytes.Buffer)
			runGuest(vmm.Config{
				Devices: []virtio.DeviceConfig{
					&virtio.ConsoleDevice{
						Out: out,
					},
				},
			})

			if !strings.Contains(out.String(), "hello from the guest") {
				t.Error("the guest didn't say hello")
			}
		},

		Guest: func(t *testing.T) {
			if _, err := fmt.Fprintln(os.Stdout, "hello from the guest"); err != nil {
				t.Fatal(err)
			}
		},
	}.Run(t)
}
