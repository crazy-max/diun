package registry

import (
	"path"

	regplatform "github.com/regclient/regclient/types/platform"
)

func TargetPlatform(os, arch, variant string) (regplatform.Platform, error) {
	local := regplatform.Local()
	if os == "" && arch == "" && variant == "" {
		return local, nil
	}
	if os == "" {
		os = local.OS
	}
	if arch == "" {
		arch = local.Architecture
	}
	if variant == "" && arch == local.Architecture {
		variant = local.Variant
	}
	return regplatform.Parse(path.Join(os, arch, variant))
}
