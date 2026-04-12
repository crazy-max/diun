//go:build !386 && !amd64 && !amd64p32 && !arm && !arm64

package platform

func lookupCPUVariant() string {
	return ""
}
