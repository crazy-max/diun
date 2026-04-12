//go:build !windows

package conffile

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

const (
	appDirEnv = "XDG_CONFIG_HOME"
	homeEnv   = "HOME"
)

func appDir() string {
	appDir := os.Getenv(appDirEnv)
	if appDir == "" {
		home := homeDir()
		appDir = filepath.Join(home, ".config")
	}
	return appDir
}

func getFileOwner(stat fs.FileInfo) (int, int, error) {
	var uid, gid int
	if sysstat, ok := stat.Sys().(*syscall.Stat_t); ok {
		uid = int(sysstat.Uid)
		gid = int(sysstat.Gid)
	}
	return uid, gid, nil
}

func osString(unix, _ string) string {
	return unix
}
