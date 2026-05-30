//go:build windows

package containerd

import (
	"context"
	"net"
	"path/filepath"

	winio "github.com/Microsoft/go-winio"
)

const defaultEndpoint = `\\.\pipe\containerd-containerd`

func dialAddress(endpoint string) string {
	return "npipe://" + filepath.ToSlash(endpoint)
}

func contextDialer(ctx context.Context, address string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, normalizeEndpoint(filepath.ToSlash(address)))
}
