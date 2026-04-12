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
