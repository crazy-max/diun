//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/require"
)

func runApplicationFixture(t *testing.T, fixtureName, imageName, tag string) {
	t.Helper()
	fixture := filepath.Join("/src/test/fixtures", fixtureName)
	data, err := os.ReadFile(filepath.Join(fixture, "diun.yml"))
	require.NoError(t, err)
	filename := filepath.Join(t.TempDir(), "diun.yml")
	require.NoError(t, os.WriteFile(filename, []byte(strings.ReplaceAll(string(data), "/mount", filepath.Join(fixture, "mount"))), 0o600))
	t.Setenv("DIUN_DB_PATH", filepath.Join(t.TempDir(), "diun.db"))
	cfg, err := config.Load(filename)
	require.NoError(t, err)
	if cfg.Providers.Docker != nil {
		cfg.Providers.Docker.Endpoint = os.Getenv("DOCKER_HOST")
		cfg.Providers.Docker.TLSVerify = new(false)
	}
	diun, err := app.New(model.Meta{ID: "diun", Name: "Diun", Version: "integration"}, cfg, ":42286")
	require.NoError(t, err)
	ctx, cancel := context.WithCancelCause(t.Context())
	done := make(chan error, 1)
	go func() { done <- diun.Start(ctx) }()
	t.Cleanup(func() {
		cancel(nil)
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(30 * time.Second):
			t.Error("application did not stop")
		}
	})
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeoutCause(t.Context(), 5*time.Second, errors.New("image list timed out"))
		defer cancel()
		output, err := exec.CommandContext(ctx, "diun", "image", "list", "--raw").Output()
		if err != nil {
			return false
		}
		var result struct {
			Images []struct {
				Name   string
				Latest struct{ Tag string }
			}
		}
		if json.Unmarshal(output, &result) != nil {
			return false
		}
		for _, image := range result.Images {
			if image.Name == imageName && (tag == "" || image.Latest.Tag == tag) {
				return true
			}
		}
		return false
	}, 4*time.Minute, time.Second, "expected %s:%s in the database", imageName, tag)
	runCommand(t, "diun", "healthcheck", "--raw")
}

func runCommand(t *testing.T, name string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeoutCause(context.Background(), 2*time.Minute, errors.New("fixture command timed out"))
	defer cancel()
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	require.NoError(t, err, "%s %v: %s", name, args, output)
}
