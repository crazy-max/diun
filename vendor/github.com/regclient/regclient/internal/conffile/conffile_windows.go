//go:build windows

package conffile

import (
	"io/fs"
	"os"
	"path/filepath"
)

const (
	appDirEnv = "APPDATA"
	homeEnv   = "USERPROFILE"
)

func appDir() string {
	appDir := os.Getenv(appDirEnv)
	if appDir == "" {
		home := homeDir()
		appDir = filepath.Join(home, "AppData")
	}
	return appDir
}

func getFileOwner(_ fs.FileInfo) (int, int, error) {
	return 0, 0, nil
}

func osString(_, win string) string {
	return win
}
