//go:build integration

package integration

import "testing"

func TestFileProvider(t *testing.T) {
	t.Run("file1", func(t *testing.T) {
		runApplicationFixture(t, "file1", "docker.io/library/alpine", "")
	})
	t.Run("file-artifact", func(t *testing.T) {
		runApplicationFixture(t, "file-artifact", "ghcr.io/crazy-max/7zip", "")
	})
}
