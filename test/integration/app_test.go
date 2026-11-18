//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	cfg.Metrics.Enabled = new(true)
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
	// A saved image does not mean the startup watch has finished its other jobs.
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:9090"+cfg.Metrics.Path, nil)
	require.NoError(t, err)
	if cfg.Metrics.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Metrics.Token)
	}
	require.Eventually(t, func() bool {
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		return err == nil && resp.StatusCode == http.StatusOK && strings.Contains(string(body), "diun_watch_runs_total 1\n")
	}, 4*time.Minute, time.Second, "startup watch did not finish")

	listCtx, listCancel := context.WithTimeoutCause(t.Context(), 5*time.Second, errors.New("image list timed out"))
	defer listCancel()
	output, err := exec.CommandContext(listCtx, "diun", "image", "list", "--raw").Output()
	require.NoError(t, err)
	var result struct {
		Images []struct {
			Name   string
			Latest struct{ Tag string }
		}
	}
	require.NoError(t, json.Unmarshal(output, &result))
	var found bool
	for _, image := range result.Images {
		if image.Name == imageName && (tag == "" || image.Latest.Tag == tag) {
			found = true
			break
		}
	}
	require.True(t, found, "expected %s:%s in the database", imageName, tag)
	runCommand(t, "diun", "healthcheck", "--raw")
}

func runCommand(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeoutCause(context.Background(), 2*time.Minute, errors.New("fixture command timed out"))
	defer cancel()
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	require.NoError(t, err, "%s %v: %s", name, args, output)
	return output
}
