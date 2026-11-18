//go:build integration

package integration

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/opencontainers/go-digest"
	imgspec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runtimeImage struct {
	index    digest.Digest
	manifest digest.Digest
}

func requireLocalDigestFixture(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_LOCAL_DIGEST") != "1" {
		t.Skip("run with test/localdigest.compose.yaml")
	}
}

// A static sleeper avoids public base images and gives all three runtimes the
// same runnable layer. Each revision has a different config and image index.
func runnablePublisher(t *testing.T, f *registryFixture) func(string) runtimeImage {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(source, []byte(`package main
import ("os"; "os/signal"; "syscall")
func main() {
 c := make(chan os.Signal, 1)
 signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
 <-c
}
`), 0o600))
	binary := filepath.Join(dir, "sleeper")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", binary, source)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
	data, err := os.ReadFile(binary)
	require.NoError(t, err)
	var layer bytes.Buffer
	tw := tar.NewWriter(&layer)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "sleeper", Mode: 0o755, Size: int64(len(data))}))
	_, err = tw.Write(data)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	layerDigest := uploadRuntimeBlob(t, f, layer.Bytes())
	return func(revision string) runtimeImage {
		config, err := json.Marshal(map[string]any{
			"architecture": runtime.GOARCH,
			"os":           "linux",
			"config":       map[string]any{"Entrypoint": []string{"/sleeper"}, "Labels": map[string]string{"revision": revision}},
			"rootfs":       map[string]any{"type": "layers", "diff_ids": []digest.Digest{layerDigest}},
		})
		require.NoError(t, err)
		configDigest := uploadRuntimeBlob(t, f, config)
		manifest, err := json.Marshal(map[string]any{
			"schemaVersion": 2,
			"mediaType":     imgspec.MediaTypeImageManifest,
			"config":        imgspec.Descriptor{MediaType: imgspec.MediaTypeImageConfig, Digest: configDigest, Size: int64(len(config))},
			"layers":        []imgspec.Descriptor{{MediaType: imgspec.MediaTypeImageLayer, Digest: layerDigest, Size: int64(layer.Len())}},
		})
		require.NoError(t, err)
		manifestDigest := digest.FromBytes(manifest)
		f.request(t, http.MethodPut, f.url+"/v2/"+f.repo+"/manifests/"+manifestDigest.String(), imgspec.MediaTypeImageManifest, manifest, http.StatusCreated)
		index, err := json.Marshal(map[string]any{
			"schemaVersion": 2,
			"mediaType":     imgspec.MediaTypeImageIndex,
			"manifests":     []imgspec.Descriptor{{MediaType: imgspec.MediaTypeImageManifest, Digest: manifestDigest, Size: int64(len(manifest)), Platform: &imgspec.Platform{OS: "linux", Architecture: runtime.GOARCH}}},
		})
		require.NoError(t, err)
		f.request(t, http.MethodPut, f.url+"/v2/"+f.repo+"/manifests/latest", imgspec.MediaTypeImageIndex, index, http.StatusCreated)
		return runtimeImage{index: digest.FromBytes(index), manifest: manifestDigest}
	}
}

func uploadRuntimeBlob(t *testing.T, f *registryFixture, blob []byte) digest.Digest {
	t.Helper()
	dgst := digest.FromBytes(blob)
	headers := f.request(t, http.MethodPost, f.url+"/v2/"+f.repo+"/blobs/uploads/", "", nil, http.StatusAccepted)
	location := headers.Get("Location")
	if strings.HasPrefix(location, "/") {
		location = f.url + location
	}
	separator := "?"
	if strings.Contains(location, "?") {
		separator = "&"
	}
	f.request(t, http.MethodPut, location+separator+"digest="+dgst.String(), "application/octet-stream", blob, http.StatusCreated)
	return dgst
}

func startDigestWatcher(t *testing.T, f *registryFixture, providers *model.Providers) *app.Diun {
	t.Helper()
	f.cfg.Providers = providers
	f.cfg.Watch.Schedule = "@every 24h"
	f.cfg.Watch.Jitter = new(time.Duration(0))
	f.cfg.Metrics = (&model.Metrics{}).GetDefaults()
	f.cfg.Metrics.Enabled = new(true)
	f.cfg.Metrics.Addr = "127.0.0.1:9090"
	diun, err := app.New(model.Meta{Version: "integration"}, f.cfg, "127.0.0.1:0")
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
			t.Error("watcher did not stop")
		}
	})
	require.Eventually(t, func() bool {
		body, err := scrapeDigestMetrics(t.Context())
		return err == nil && strings.Contains(body, "diun_watch_runs_total 1\n")
	}, time.Minute, 200*time.Millisecond, "startup watch did not finish")
	return diun
}

func scrapeDigestMetrics(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:9090/metrics", nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// Count watch runs as well as the gauge, proving a rollout clears the metric
// without silently relying on another registry poll.
func expectDigestMetric(t *testing.T, provider string, want, runs int) {
	t.Helper()
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		body, err := scrapeDigestMetrics(t.Context())
		if !assert.NoError(c, err) {
			return
		}
		assert.Contains(c, body, fmt.Sprintf("diun_watch_runs_total %d\n", runs))
		found := false
		for line := range strings.SplitSeq(body, "\n") {
			if strings.HasPrefix(line, "diun_image_local_digest_mismatch{") && strings.Contains(line, `provider="`+provider+`"`) {
				assert.True(c, strings.HasSuffix(line, fmt.Sprintf(" %d", want)), "%s", line)
				found = true
			}
		}
		assert.True(c, found, "comparison series missing for %s", provider)
	}, time.Minute, 200*time.Millisecond, "provider %s mismatch=%d at run %d", provider, want, runs)
}
