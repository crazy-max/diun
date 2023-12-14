//go:build !windows
// +build !windows

package script

import (
	"os/exec"
)

func setSysProcAttr(_ *exec.Cmd) {
}
