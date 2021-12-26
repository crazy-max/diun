//go:build !windows
// +build !windows

package utl

import (
	"golang.org/x/sys/unix"
)

const (
	SIGTERM = unix.SIGTERM
)
