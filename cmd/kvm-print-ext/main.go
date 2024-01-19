// kvm-print-ext prints information about the KVM API and extensions.
package main

import (
	"fmt"
	"os"

	"github.com/c35s/hype/kvm"
)

func main() {
	sys, err := os.Open("/dev/kvm")
	if err != nil {
		panic(err)
	}

	defer sys.Close()

	version, err := kvm.GetAPIVersion(sys)
	if err != nil {
		panic(err)
	}

	fmt.Printf("KVM API version: %d\n", version)

	fmt.Println("\n# extensions")
	for _, c := range kvm.AllCaps() {
		v, err := kvm.CheckExtension(sys, c)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%v: %v\n", c, v)
	}
}
