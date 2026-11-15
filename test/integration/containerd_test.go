//go:build integration

package integration

import "testing"

func TestContainerdProvider(t *testing.T) {
	runApplicationFixture(t, "containerd1", "docker.io/library/busybox", "1.36.1")
}
