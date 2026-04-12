package platform

import (
	"strconv"
	"strings"
)

type compare struct {
	host Platform
}

type CompareOpts func(*compare)

// NewCompare is used to compare multiple target entries to a host value.
func NewCompare(host Platform, opts ...CompareOpts) *compare {
	(&host).normalize()
	c := compare{
		host: host,
	}
	for _, optFn := range opts {
		optFn(&c)
	}
	return &c
}

// Better returns true when the target is compatible and a better match than the previous platform.
// The previous platform value may be the zero value when no previous match has been found.
func (c *compare) Better(target, prev Platform) bool {
	if !Compatible(c.host, target) {
		return false
	}
	(&target).normalize()
	(&prev).normalize()
	if prev.OS != target.OS {
		if target.OS == c.host.OS {
			return true
		} else if prev.OS == c.host.OS {
			return false
		}
	}
	if prev.Architecture != target.Architecture {
		if target.Architecture == c.host.Architecture {
			return true
		} else if prev.Architecture == c.host.Architecture {
			return false
		}
	}
	if prev.Variant != target.Variant {
		if target.Variant == c.host.Variant {
			return true
		} else if prev.Variant == c.host.Variant {
			return false
		}
		pV := variantVer(prev.Variant)
		tV := variantVer(target.Variant)
		if tV > pV {
			return true
		} else if tV < pV {
			return false
		}
	}
	if prev.OSVersion != target.OSVersion {
		if target.OSVersion == c.host.OSVersion {
			return true
		} else if prev.OSVersion == c.host.OSVersion {
			return false
		}
		cmp := semverCmp(prev.OSVersion, target.OSVersion)
		if cmp != 0 {
			return cmp < 0
		}
	}
	return false
}

// Compatible indicates if a host can run a specified target platform image.
// This accounts for Docker Desktop for Mac and Windows using a Linux VM.
func (c *compare) Compatible(target Platform) bool {
	(&target).normalize()
	if c.host.OS == "linux" || c.host.OS == "freebsd" {
		return c.host.OS == target.OS && c.host.Architecture == target.Architecture &&
			variantCompatible(c.host.Variant, target.Variant)
	} else if c.host.OS == "windows" {
		if target.OS == "windows" {
			return c.host.Architecture == target.Architecture &&
				variantCompatible(c.host.Variant, target.Variant) &&
				osVerCompatible(c.host.OSVersion, target.OSVersion)
		} else if target.OS == "linux" {
			return c.host.Architecture == target.Architecture &&
				variantCompatible(c.host.Variant, target.Variant)
		}
		return false
	} else if c.host.OS == "darwin" {
		return (target.OS == "darwin" || target.OS == "linux") &&
			c.host.Architecture == target.Architecture &&
			variantCompatible(c.host.Variant, target.Variant)
	} else {
		return c.host.OS == target.OS && c.host.Architecture == target.Architecture &&
			variantCompatible(c.host.Variant, target.Variant) &&
			c.host.OSVersion == target.OSVersion &&
			strSliceEq(c.host.OSFeatures, target.OSFeatures) &&
			strSliceEq(c.host.Features, target.Features)
	}
}

// Match indicates if two platforms are the same.
func (c *compare) Match(target Platform) bool {
	(&target).normalize()
	if c.host.OS != target.OS {
		return false
	}
	if c.host.OS == "linux" || c.host.OS == "freebsd" {
		return c.host.Architecture == target.Architecture && c.host.Variant == target.Variant
	} else if c.host.OS == "windows" {
		return c.host.Architecture == target.Architecture && c.host.Variant == target.Variant &&
			osVerSemver(c.host.OSVersion) == osVerSemver(target.OSVersion)
	} else {
		return c.host.Architecture == target.Architecture &&
			c.host.Variant == target.Variant &&
			c.host.OSVersion == target.OSVersion &&
			strSliceEq(c.host.OSFeatures, target.OSFeatures) &&
			strSliceEq(c.host.Features, target.Features)
	}
}

// Compatible indicates if a host can run a specified target platform image.
// This accounts for Docker Desktop for Mac and Windows using a Linux VM.
func Compatible(host, target Platform) bool {
	comp := NewCompare(host)
	return comp.Compatible(target)
}

// Match indicates if two platforms are the same.
func Match(a, b Platform) bool {
	comp := NewCompare(a)
	return comp.Match(b)
}

func osVerCompatible(host, target string) bool {
	if host == "" {
		return true
	}
	vHost := osVerSemver(host)
	vTarget := osVerSemver(target)
	return vHost == vTarget
}

func osVerSemver(platVer string) string {
	verParts := strings.Split(platVer, ".")
	if len(verParts) < 4 {
		return platVer
	}
	return strings.Join(verParts[0:3], ".")
}

// return: -1 if a<b, 0 if a==b, 1 if a>b
func semverCmp(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := range aParts {
		if len(bParts) < i+1 {
			return 1
		}
		aInt, aErr := strconv.Atoi(aParts[i])
		bInt, bErr := strconv.Atoi(bParts[i])
		if aErr != nil {
			if bErr != nil {
				return 0
			}
			return -1
		}
		if bErr != nil {
			return 1
		}
		if aInt < bInt {
			return -1
		}
		if aInt > bInt {
			return 1
		}
	}
	return 0
}

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func variantCompatible(host, target string) bool {
	vHost := variantVer(host)
	vTarget := variantVer(target)
	if vHost >= vTarget || (vHost == 1 && target == "") || (host == "" && vTarget == 1) {
		return true
	}
	return false
}

func variantVer(v string) int {
	v = strings.TrimPrefix(v, "v")
	ver, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return ver
}
