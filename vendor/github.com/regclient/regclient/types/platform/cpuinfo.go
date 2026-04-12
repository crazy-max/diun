// Related implementations:
// <https://golang.org/x/sys/cpu>
// <https://github.com/klauspost/cpuid>
// <https://github.com/containerd/platforms>
// <https://github.com/tonistiigi/go-archvariant>
// <https://tip.golang.org/wiki/MinimumRequirements#microarchitecture-support>

package platform

import (
	"runtime"
	"sync"
)

// cpuVariantValue is the variant of the local CPU architecture.
// For example on ARM, v7 and v8. And on AMD64, v1 - v4.
// Don't use this value directly; call cpuVariant() instead.
var cpuVariantValue string

var cpuVariantOnce sync.Once

func cpuVariant() string {
	cpuVariantOnce.Do(func() {
		switch runtime.GOARCH {
		case "amd64", "arm", "arm64":
			cpuVariantValue = lookupCPUVariant()
		}
	})
	return cpuVariantValue
}
