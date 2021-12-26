//go:build windows
// +build windows

package utl

import (
	"golang.org/x/sys/windows"
)

const (
	SIGTERM = windows.SIGTERM
)
