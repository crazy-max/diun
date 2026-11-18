//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Each fixture uses a separate repository and database, including across reruns.
type registryFixture struct {
	url     string
	repo    string
	cfg     *config.Config
	mu      sync.Mutex
	reports []map[string]any
}

func newRegistryFixture(t *testing.T) *registryFixture {
	t.Helper()
	endpoint := os.Getenv("TEST_REGISTRY")
	if endpoint == "" {
		t.Skip("TEST_REGISTRY is unset; run with test/registry.compose.yaml")
	}
	f := &registryFixture{
		url:  strings.TrimRight(endpoint, "/"),
		repo: fmt.Sprintf("diun-%d", time.Now().UnixNano()),
	}
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("X-Diun-Test") != "webhook" {
			t.Errorf("unexpected webhook method or headers: %s %v", r.Method, r.Header)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		f.mu.Lock()
		f.reports = append(f.reports, payload)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(receiver.Close)
	f.cfg = &config.Config{
		Db:        &model.Db{Path: filepath.Join(t.TempDir(), "diun.db")},
		Watch:     (&model.Watch{}).GetDefaults(),
		Defaults:  (&model.Defaults{}).GetDefaults(),
		Providers: &model.Providers{File: &model.PrdFile{Filename: filepath.Join(t.TempDir(), "images.yml")}},
		RegOpts: model.RegOpts{
			{
				Name:        "local",
				Selector:    model.RegOptSelectorName,
				InsecureTLS: new(true),
				Timeout:     new(5 * time.Second),
			},
		},
		Notif: &model.Notif{Webhook: &model.NotifWebhook{
			Endpoint: receiver.URL,
			Method:   http.MethodPost,
			Headers:  map[string]string{"X-Diun-Test": "webhook"},
			Timeout:  new(5 * time.Second),
		}},
	}
	return f
}

func (f *registryFixture) image(tag string) model.Image {
	return model.Image{
		Name:     strings.TrimPrefix(f.url, "http://") + "/" + f.repo + ":" + tag,
		RegOpt:   "local",
		Platform: model.ImagePlatform{OS: "linux", Arch: "amd64"},
	}
}

func (f *registryFixture) request(t *testing.T, method, endpoint, contentType string, body []byte, status int) http.Header {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, endpoint, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode, "%s %s: %s", method, endpoint, data)
	return resp.Header
}

func (f *registryFixture) push(t *testing.T, tag, revision string) digest.Digest {
	t.Helper()
	configBlob, err := json.Marshal(map[string]any{
		"architecture": "amd64",
		"os":           "linux",
		"created":      "2026-01-01T00:00:00Z",
		"config": map[string]any{"Labels": map[string]string{
			"org.opencontainers.image.revision": revision,
			"org.opencontainers.image.url":      "https://example.com/image",
		}},
		"rootfs": map[string]any{"type": "layers", "diff_ids": []string{}},
	})
	require.NoError(t, err)
	configDigest := digest.FromBytes(configBlob)
	headers := f.request(t, http.MethodPost, f.url+"/v2/"+f.repo+"/blobs/uploads/", "", nil, http.StatusAccepted)
	location := headers.Get("Location")
	if strings.HasPrefix(location, "/") {
		location = f.url + location
	}
	separator := "?"
	if strings.Contains(location, "?") {
		separator = "&"
	}
	f.request(t, http.MethodPut, location+separator+"digest="+configDigest.String(), "application/octet-stream", configBlob, http.StatusCreated)
	manifest, err := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.manifest.v1+json",
		"config": map[string]any{
			"mediaType": "application/vnd.oci.image.config.v1+json",
			"digest":    configDigest,
			"size":      len(configBlob),
		},
		"layers": []any{},
	})
	require.NoError(t, err)
	f.request(t, http.MethodPut, f.url+"/v2/"+f.repo+"/manifests/"+tag, "application/vnd.oci.image.manifest.v1+json", manifest, http.StatusCreated)
	return digest.FromBytes(manifest)
}

// With no schedule, Start returns after the scan and closes the database.
// Every scan therefore also checks persistence across application restarts.
func (f *registryFixture) scan(t *testing.T, images ...model.Image) ([]map[string]any, []registry.Manifest) {
	t.Helper()
	data, err := yaml.Marshal(images)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(f.cfg.Providers.File.Filename, data, 0o600))
	diun, err := app.New(model.Meta{Version: "integration", Hostname: "registry-test"}, f.cfg, "127.0.0.1:0")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	require.NoError(t, diun.Start(ctx))
	require.NoError(t, ctx.Err(), "scan exceeded deadline")
	f.mu.Lock()
	reports := f.reports
	f.reports = nil
	f.mu.Unlock()
	store, err := db.New(*f.cfg.Db)
	require.NoError(t, err)
	defer store.Close()
	manifests, err := store.ListManifest()
	require.NoError(t, err)
	return reports, manifests
}

func TestRegistryWebhookLifecycle(t *testing.T) {
	for _, compareDigest := range []bool{true, false} {
		t.Run(fmt.Sprintf("compareDigest=%t", compareDigest), func(t *testing.T) {
			f := newRegistryFixture(t)
			f.cfg.Watch.CompareDigest = new(compareDigest)
			image := f.image("latest")
			image.Metadata = map[string]string{"service": "example"}
			first := f.push(t, "latest", "first")
			reports, saved := f.scan(t, image)
			require.Empty(t, reports)
			require.Len(t, saved, 1)
			require.Equal(t, first, saved[0].Digest)

			second := f.push(t, "latest", "second")
			reports, saved = f.scan(t, image)
			require.Len(t, reports, 1)
			require.Equal(t, "update", reports[0]["status"])
			require.Equal(t, image.Name, reports[0]["image"])
			require.Equal(t, second.String(), reports[0]["digest"])
			require.Equal(t, "file", reports[0]["provider"])
			require.Equal(t, "integration", reports[0]["diun_version"])
			require.Equal(t, "registry-test", reports[0]["hostname"])
			require.Equal(t, "linux/amd64", reports[0]["platform"])
			require.Equal(t, "2026-01-01T00:00:00Z", reports[0]["created"])
			require.Equal(t, "application/vnd.oci.image.manifest.v1+json", reports[0]["mime_type"])
			require.Equal(t, "https://example.com/image", reports[0]["hub_link"])
			require.Equal(t, map[string]any{"service": "example"}, reports[0]["metadata"])
			require.NotContains(t, reports[0], "state")
			require.Len(t, saved, 1)
			require.Equal(t, second, saved[0].Digest)

			reports, _ = f.scan(t, image)
			require.Empty(t, reports)
			image.NotifyOn = []model.NotifyOn{model.NotifyOnUnchange}
			image.HubLink = "https://example.com/override"
			for range 2 {
				reports, _ = f.scan(t, image)
				require.Len(t, reports, 1)
				require.Equal(t, "unchange", reports[0]["status"])
				require.Equal(t, "https://example.com/override", reports[0]["hub_link"])
			}

			third := f.push(t, "latest", "third")
			reports, saved = f.scan(t, image)
			require.Empty(t, reports, "update is filtered out")
			require.Equal(t, third, saved[0].Digest)
			image.NotifyOn = model.NotifyOnDefaults
			reports, _ = f.scan(t, image)
			require.Empty(t, reports, "suppressed update was still persisted")

			f.request(t, http.MethodDelete, f.url+"/v2/"+f.repo+"/manifests/"+third.String(), "", nil, http.StatusAccepted)
			reports, saved = f.scan(t, image)
			require.Empty(t, reports)
			require.Equal(t, third, saved[0].Digest, "registry failure preserves cache")
			f.push(t, "latest", "fourth")
			reports, _ = f.scan(t, image)
			require.Len(t, reports, 1)
			require.Equal(t, "update", reports[0]["status"])
		})
	}
}

func TestRegistryFirstCheckNotifications(t *testing.T) {
	for _, tt := range []struct {
		name       string
		firstCheck bool
		notifyOn   []model.NotifyOn
		want       int
	}{
		{
			name: "disabled",
		},
		{
			name:       "enabled",
			firstCheck: true,
			want:       1,
		},
		{
			name:       "enabled but filtered",
			firstCheck: true,
			notifyOn:   []model.NotifyOn{model.NotifyOnUpdate},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newRegistryFixture(t)
			f.cfg.Watch.FirstCheckNotif = new(tt.firstCheck)
			dgst := f.push(t, "latest", "first")
			image := f.image("latest")
			image.NotifyOn = tt.notifyOn
			reports, saved := f.scan(t, image)
			require.Len(t, reports, tt.want)
			if tt.want != 0 {
				require.Equal(t, "new", reports[0]["status"])
			}
			require.Len(t, saved, 1)
			require.Equal(t, dgst, saved[0].Digest)
			reports, _ = f.scan(t, image)
			require.Empty(t, reports, "restart must not repeat first-check notification")
		})
	}
}

func TestRegistryWebhookFailure(t *testing.T) {
	f := newRegistryFixture(t)
	f.cfg.Watch.FirstCheckNotif = new(true)
	image := f.image("latest")
	f.push(t, "latest", "first")
	reports, _ := f.scan(t, image)
	require.Len(t, reports, 1)

	attempts := make(chan struct{}, 8)
	failedReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts <- struct{}{}
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer failedReceiver.Close()
	endpoint := f.cfg.Notif.Webhook.Endpoint
	f.cfg.Notif.Webhook.Endpoint = failedReceiver.URL
	updated := f.push(t, "latest", "second")
	reports, saved := f.scan(t, image)
	require.Empty(t, reports)
	require.Len(t, attempts, 1)
	require.Len(t, saved, 1)
	require.Equal(t, updated, saved[0].Digest)

	// Delivery failure does not roll back the scan or queue a retry.
	f.cfg.Notif.Webhook.Endpoint = endpoint
	reports, _ = f.scan(t, image)
	require.Empty(t, reports)
	f.push(t, "latest", "third")
	reports, _ = f.scan(t, image)
	require.Len(t, reports, 1)
	require.Equal(t, "update", reports[0]["status"])
}

func TestRegistryMissingImageDoesNotBlockNotifications(t *testing.T) {
	f := newRegistryFixture(t)
	f.cfg.Watch.FirstCheckNotif = new(true)
	f.push(t, "latest", "first")
	image := f.image("latest")
	reports, saved := f.scan(t, f.image("missing"), image)
	require.Len(t, reports, 1)
	require.Equal(t, image.Name, reports[0]["image"])
	require.Equal(t, "new", reports[0]["status"])
	require.Len(t, saved, 1)
	require.Equal(t, "latest", saved[0].Tag)
}

func TestRegistryFilteredImages(t *testing.T) {
	for _, tt := range []struct {
		name    string
		include []string
		exclude []string
	}{
		{
			name:    "include mismatch",
			include: []string{"^stable$"},
		},
		{
			name:    "exclusion wins",
			include: []string{"^latest$"},
			exclude: []string{"^latest$"},
		},
		{
			name:    "invalid include regex",
			include: []string{"["},
		},
		{
			name:    "invalid exclude regex",
			exclude: []string{"["},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newRegistryFixture(t)
			f.cfg.Watch.FirstCheckNotif = new(true)
			f.push(t, "latest", "first")
			image := f.image("latest")
			image.IncludeTags = tt.include
			image.ExcludeTags = tt.exclude
			reports, saved := f.scan(t, image)
			require.Empty(t, reports)
			require.Empty(t, saved)
		})
	}
}

func TestRegistryTagDiscovery(t *testing.T) {
	for _, tt := range []struct {
		name    string
		mode    registry.WatchMode
		include []string
		exclude []string
		max     int
		want    []string
	}{
		{
			name: "tag only",
			mode: registry.WatchModeTag,
		},
		{
			name: "all",
			mode: registry.WatchModeAll,
			want: []string{"0.9.0", "1.1.0", "2.0.0", "2.1.0-rc.1", "edge"},
		},
		{
			name: "semver",
			mode: registry.WatchModeSemver,
			want: []string{"1.1.0", "2.0.0"},
		},
		{
			name: "semver prereleases",
			mode: registry.WatchModeSemverPrereleases,
			want: []string{"1.1.0", "2.0.0", "2.1.0-rc.1"},
		},
		{
			name:    "include",
			mode:    registry.WatchModeAll,
			include: []string{"^1[.]"},
			want:    []string{"1.1.0"},
		},
		{
			name:    "exclude wins",
			mode:    registry.WatchModeAll,
			include: []string{"^1[.]"},
			exclude: []string{"^1[.]1[.]0$"},
		},
		{
			name: "max tags",
			mode: registry.WatchModeSemver,
			max:  1,
			want: []string{"2.0.0"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newRegistryFixture(t)
			f.push(t, "1.0.0", "baseline")
			image := f.image("1.0.0")
			image.WatchMode = tt.mode
			image.IncludeTags = tt.include
			image.ExcludeTags = tt.exclude
			image.MaxTags = tt.max
			image.SortTags = registry.SortTagSemver
			reports, saved := f.scan(t, image)
			require.Empty(t, reports)
			require.Len(t, saved, 1)
			for _, tag := range []string{"0.9.0", "1.1.0", "2.0.0", "2.1.0-rc.1", "edge", "sha256-" + strings.Repeat("a", 64) + ".sig"} {
				f.push(t, tag, tag)
			}
			reports, saved = f.scan(t, image)
			require.Len(t, reports, len(tt.want))
			var got []string
			for _, report := range reports {
				require.Equal(t, "new", report["status"])
				got = append(got, strings.TrimPrefix(report["image"].(string), strings.TrimSuffix(image.Name, "1.0.0")))
			}
			require.ElementsMatch(t, tt.want, got)
			require.Len(t, saved, len(tt.want)+1)
			reports, _ = f.scan(t, image)
			require.Empty(t, reports, "discovered tags must not notify twice")
		})
	}
}
