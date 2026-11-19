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

//go:build windows

package platform

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"
)

// Local retrieves the local platform details
func Local() Platform {
	major, minor, build := windows.RtlGetNtVersionNumbers()
	plat := Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Variant:      cpuVariant(),
		OSVersion:    fmt.Sprintf("%d.%d.%d", major, minor, build),
	}
	plat.normalize()
	return plat
}
