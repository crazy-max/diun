package exhttp

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func isWindowsNetworkError(errno syscall.Errno) bool {
	switch errno {
	case windows.WSAENETDOWN,
		windows.WSAENETUNREACH,
		windows.WSAENETRESET,
		windows.WSAECONNABORTED,
		windows.WSAECONNRESET,
		windows.WSAENOBUFS,
		windows.WSAETIMEDOUT,
		windows.WSAECONNREFUSED,
		windows.WSAEHOSTDOWN,
		windows.WSAEHOSTUNREACH,
		windows.WSAESHUTDOWN:
		return true
	default:
		return false
	}
}
