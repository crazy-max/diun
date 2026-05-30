//go:build !windows

package containerd

import (
	"context"
	"net"
)

func dialAddress(endpoint string) string {
	return "unix://" + endpoint
}

func contextDialer(ctx context.Context, address string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "unix", normalizeEndpoint(address))
}
