package arch

import (
	"fmt"
	"os"
	"strings"

	"github.com/c35s/hype/kvm"
)

// requiredCaps are the KVM extensions required for all architectures.
// See archCaps for required arch-specific extensions.
var requiredCaps = []kvm.Cap{
	kvm.CapIRQChip,
	kvm.CapHLT,
	kvm.CapUserMemory,
	kvm.CapIRQFD,
	kvm.CapCheckExtensionVM,
	kvm.CapImmediateExit,
}

// ValidateKVM returns an error if KVM doesn't support the required extensions.
func ValidateKVM(kfd *os.File) error {
	version, err := kvm.GetAPIVersion(kfd)
	if err != nil {
		return err
	}

	if version != kvm.StableAPIVersion {
		return fmt.Errorf("unstable API version: %d != %d", version, kvm.StableAPIVersion)
	}

	caps := append([]kvm.Cap(nil), requiredCaps...)
	caps = append(caps, archCaps...)

	var missing []kvm.Cap
	for _, cap := range caps {
		val, err := kvm.CheckExtension(kfd, cap)
		if err != nil {
			return err
		}

		if val < 1 {
			missing = append(missing, cap)
		}
	}

	if len(missing) > 0 {
		var names []string
		for _, cap := range missing {
			names = append(names, cap.String())
		}

		return fmt.Errorf("missing %s", strings.Join(names, ","))
	}

	return nil
}
