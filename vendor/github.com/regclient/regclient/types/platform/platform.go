// Package platform handles the parsing and comparing of the image platform (e.g. linux/amd64)
package platform

// Some of the code in the package and all of the inspiration for this comes from <https://github.com/containerd/containerd>.
// Their license is included here:
/*
   Copyright The containerd Authors.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/regclient/regclient/internal/strparse"
	"github.com/regclient/regclient/types/errs"
)

var partRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Platform specifies a platform where a particular image manifest is applicable.
type Platform struct {
	// Architecture field specifies the CPU architecture, for example `amd64` or `ppc64`.
	Architecture string `json:"architecture"`

	// OS specifies the operating system, for example `linux` or `windows`.
	OS string `json:"os"`

	// OSVersion is an optional field specifying the operating system version, for example `10.0.10586`.
	OSVersion string `json:"os.version,omitempty"`

	// OSFeatures is an optional field specifying an array of strings, each listing a required OS feature (for example on Windows `win32k`).
	OSFeatures []string `json:"os.features,omitempty"`

	// Variant is an optional field specifying a variant of the CPU, for example `ppc64le` to specify a little-endian version of a PowerPC CPU.
	Variant string `json:"variant,omitempty"`

	// Features is an optional field specifying an array of strings, each listing a required CPU feature (for example `sse4` or `aes`).
	Features []string `json:"features,omitempty"`
}

// String outputs the platform in the <os>/<arch>/<variant> notation
func (p Platform) String() string {
	(&p).normalize()
	if p.OS == "" {
		return "unknown"
	} else {
		return path.Join(p.OS, p.Architecture, p.Variant)
	}
}

// Parse converts a platform string into a struct
func Parse(platStr string) (Platform, error) {
	// args are a regclient specific way to extend the platform string
	platArgs := strings.SplitN(platStr, ",", 2)
	// split on slash, validate each component
	platSplit := strings.Split(platArgs[0], "/")
	for i, part := range platSplit {
		if !partRE.MatchString(part) {
			return Platform{}, fmt.Errorf("invalid platform component %s in %s%.0w", part, platStr, errs.ErrParsingFailed)
		}
		platSplit[i] = strings.ToLower(part)
	}
	plat := &Platform{}
	if len(platSplit) == 1 && knownArch(platSplit[0]) {
		// special case of architecture only
		plat.Architecture = platSplit[0]
	} else if len(platSplit) >= 1 {
		plat.OS = platSplit[0]
	}
	if len(platSplit) >= 2 {
		plat.Architecture = platSplit[1]
	}
	if len(platSplit) >= 3 {
		plat.Variant = platSplit[2]
	}
	if len(platArgs) > 1 {
		kvMap, err := strparse.SplitCSKV(platArgs[1])
		if err != nil {
			return Platform{}, fmt.Errorf("failed to split platform args in %s: %w", platStr, err)
		}
		for k, v := range kvMap {
			k := strings.TrimSpace(k)
			v := strings.TrimSpace(v)
			switch strings.ToLower(k) {
			case "osver", "osversion":
				plat.OSVersion = v
			default:
				return Platform{}, fmt.Errorf("unsupported platform arg type, %s in %s%.0w", k, platStr, errs.ErrParsingFailed)
			}
		}
	}
	// gather local platform details
	platLocal := Local()
	// normalize and extrapolate missing fields
	if platStr == "local" {
		*plat = platLocal
	} else if plat.OS == "local" || plat.OS == "" {
		plat.OS = platLocal.OS
	}
	plat.normalize()
	switch plat.OS {
	case "linux", "darwin", "windows":
		// expand short references to local platform with architecture and variant
		if Compatible(Platform{OS: platLocal.OS}, Platform{OS: plat.OS}) && len(platSplit) < 2 {
			if plat.Architecture == "" {
				plat.Architecture = platLocal.Architecture
			}
			if plat.Architecture == platLocal.Architecture && plat.Variant == "" {
				plat.Variant = platLocal.Variant
			}
		}
	}
	if plat.OS == "windows" && plat.OS == platLocal.OS && plat.Architecture == platLocal.Architecture && variantCompatible(platLocal.Variant, plat.Variant) && plat.OSVersion == "" {
		plat.OSVersion = platLocal.OSVersion
	}

	return *plat, nil
}

// knownArch is a list of known architectures that can be parsed without the OS field.
// Otherwise the OS is required.
func knownArch(arch string) bool {
	switch arch {
	case "386", "amd64", "i386", "x86_64", "x86-64",
		"arm", "armhf", "armel", "arm64", "aarch64",
		"mips", "mips64", "mips64le",
		"ppc", "ppc64", "ppc64le",
		"loong64",
		"riscv", "riscv64",
		"s390", "s390x",
		"sparc", "sparc64",
		"wasm":
		return true
	}
	return false
}

func (p *Platform) normalize() {
	switch p.OS {
	case "macos":
		p.OS = "darwin"
	}
	switch p.Architecture {
	case "i386":
		p.Architecture = "386"
		p.Variant = ""
	case "x86_64", "x86-64", "amd64":
		p.Architecture = "amd64"
		if p.Variant == "v1" {
			p.Variant = ""
		}
	case "aarch64", "arm64":
		p.Architecture = "arm64"
		switch p.Variant {
		case "8", "v8":
			p.Variant = ""
		}
	case "armhf":
		p.Architecture = "arm"
		p.Variant = "v7"
	case "armel":
		p.Architecture = "arm"
		p.Variant = "v6"
	case "arm":
		switch p.Variant {
		case "", "7":
			p.Variant = "v7"
		case "5", "6", "8":
			p.Variant = "v" + p.Variant
		}
	}
}
