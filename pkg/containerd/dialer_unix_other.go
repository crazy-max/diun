//go:build !linux && !windows

package containerd

const defaultEndpoint = "/var/run/containerd/containerd.sock"
