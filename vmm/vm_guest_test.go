//go:build guest

package vmm_test

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestConsole(t *testing.T) {
	if _, err := fmt.Fprintln(os.Stdout, "hello from the guest"); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	m.Run()

	// If the tests are running in the VM, the exit code returned by Run is
	// discarded because the kernel really doesn't like it when PID 1 exits.
	// Instead, we reboot! The host can figure out what happened by parsing
	// console output.

	if err := unix.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
		panic(err)
	}
}
