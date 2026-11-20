// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
