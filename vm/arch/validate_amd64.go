package arch

import "github.com/c35s/hype/kvm"

var archCaps = []kvm.Cap{
	kvm.CapExtCPUID,
	kvm.CapTSCDeadlineTimer,
}
