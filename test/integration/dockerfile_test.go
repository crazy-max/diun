//go:build integration

package integration

import "testing"

func TestDockerfileProvider(t *testing.T) {
	for _, fixture := range []string{"dockerfile1", "dockerfile2"} {
		t.Run(fixture, func(t *testing.T) {
			runApplicationFixture(t, fixture, "docker.io/library/alpine", "")
		})
	}
}
