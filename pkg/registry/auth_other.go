//go:build !linux

package registry

import "path/filepath"

func defaultAuthPath(home string) string {
	return filepath.Join(home, ".config", "containers", "auth.json")
}
