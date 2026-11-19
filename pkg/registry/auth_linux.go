package registry

import (
	"os"
	"path/filepath"
	"strconv"
)

func defaultAuthPath(_ string) string {
	if run := os.Getenv("XDG_RUNTIME_DIR"); run != "" {
		return filepath.Join(run, "containers", "auth.json")
	}
	return filepath.Join("/run/containers", strconv.Itoa(os.Getuid()), "auth.json")
}
