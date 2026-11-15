//go:build !windows

package exhttp

import "syscall"

func isWindowsNetworkError(syscall.Errno) bool {
	return false
}
