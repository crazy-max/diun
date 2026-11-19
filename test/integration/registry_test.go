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
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/app"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	grpcserver "github.com/crazy-max/diun/v4/internal/grpc"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pb"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json")
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
	return f.pushPlatform(t, tag, revision, model.ImagePlatform{OS: "linux", Arch: "amd64"})
}

func (f *registryFixture) pushPlatform(t *testing.T, tag, revision string, platform model.ImagePlatform) digest.Digest {
	t.Helper()
	configBlob, err := json.Marshal(map[string]any{
		"architecture": platform.Arch,
		"os":           platform.OS,
		"variant":      platform.Variant,
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

			f.pushIndex(t, "index", "index")
			reports, saved = f.scan(t, image)
			require.Len(t, reports, 1)
			require.Equal(t, "update", reports[0]["status"])
			require.NotEmpty(t, saved[0].InstanceDigest)
			f.push(t, "latest", "single again")
			reports, saved = f.scan(t, image)
			require.Len(t, reports, 1)
			require.Equal(t, "update", reports[0]["status"])
			require.Empty(t, saved[0].InstanceDigest, "a single manifest must not retain the old index instance")
			reports, _ = f.scan(t, image)
			require.Empty(t, reports)
		})
	}
}

func TestRegistryRemovalAfterTagMoves(t *testing.T) {
	for _, tt := range []struct {
		name          string
		prune         bool
		multiPlatform bool
	}{
		{
			name: "remove single manifest",
		},
		{
			name:  "prune single manifest",
			prune: true,
		},
		{
			name:          "remove multi-platform index",
			multiPlatform: true,
		},
		{
			name:          "prune multi-platform index",
			prune:         true,
			multiPlatform: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newRegistryFixture(t)
			push := func(revision string) digest.Digest {
				if !tt.multiPlatform {
					return f.push(t, "latest", revision)
				}
				f.pushIndex(t, revision, revision)
				headers := f.request(t, http.MethodHead, f.url+"/v2/"+f.repo+"/manifests/latest", "", nil, http.StatusOK)
				return digest.Digest(headers.Get("Docker-Content-Digest"))
			}
			original := push("first")
			plain := f.image("latest")
			pinned := plain
			pinned.Name += "@" + original.String()
			watches := []model.Image{pinned, plain}
			if tt.multiPlatform {
				armPinned, armPlain := pinned, plain
				armPinned.Platform.Arch = "arm64"
				armPlain.Platform.Arch = "arm64"
				watches = append(watches, armPinned, armPlain)
			}
			f.scan(t, watches...)
			current := push("second")
			reports, saved := f.scan(t, watches...)
			require.Len(t, reports, len(watches))
			require.Len(t, saved, 2)
			for _, manifest := range saved {
				require.Equal(t, current, manifest.Digest)
			}

			store, err := db.New(*f.cfg.Db)
			require.NoError(t, err)
			t.Cleanup(func() { _ = store.Close() })
			server, err := grpcserver.New("127.0.0.1:0", store, nil)
			require.NoError(t, err)
			listener, err := server.Listen()
			require.NoError(t, err)
			done := make(chan error, 1)
			go func() { done <- server.Serve(listener) }()
			t.Cleanup(server.Stop)
			conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			client := pb.NewImageServiceClient(conn)
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			if !tt.prune {
				response, err := client.ImageRemove(ctx, &pb.ImageRemoveRequest{Name: pinned.Name})
				require.NoError(t, err)
				require.Len(t, response.Manifests, 1)
				require.Equal(t, current.String(), response.Manifests[0].Digest)
			} else {
				response, err := client.ImagePrune(ctx, &pb.ImagePruneRequest{})
				require.NoError(t, err)
				require.Len(t, response.Images, 1)
				require.Len(t, response.Images[0].Manifests, 2)
			}
			images, err := store.ListImage()
			require.NoError(t, err)
			if !tt.prune {
				image, err := registry.ParseImage(registry.ParseImageOptions{Name: plain.Name})
				require.NoError(t, err)
				require.Len(t, images[image.Name()], 1)
				require.Equal(t, plain.Name, images[image.Name()][0].Reference)
			} else {
				require.Empty(t, images)
			}
			for _, watch := range watches {
				image, err := registry.ParseImage(registry.ParseImageOptions{Name: watch.Name})
				require.NoError(t, err)
				cached, err := store.GetPlatformManifest(image, watch.Platform)
				require.NoError(t, err)
				if tt.prune || watch.Name == pinned.Name {
					require.Empty(t, cached, "all platform state for removed references must be gone")
				} else {
					require.Equal(t, current, cached.Digest)
					require.Equal(t, "linux/"+watch.Platform.Arch, cached.Platform)
				}
			}
			require.NoError(t, conn.Close())
			server.Stop()
			require.NoError(t, <-done)
			require.NoError(t, store.Close())

			f.cfg.Watch.FirstCheckNotif = new(true)
			reports, _ = f.scan(t, pinned)
			require.Len(t, reports, 1)
			require.Equal(t, "new", reports[0]["status"], "the removed reference must be discovered again after restart")
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

func (f *registryFixture) pushIndex(t *testing.T, amdRevision, armRevision string) {
	t.Helper()
	var descriptors []map[string]any
	for _, instance := range []struct {
		arch     string
		revision string
	}{
		{
			arch:     "amd64",
			revision: amdRevision,
		},
		{
			arch:     "arm64",
			revision: armRevision,
		},
	} {
		if instance.revision == "" {
			continue
		}
		platform := model.ImagePlatform{OS: "linux", Arch: instance.arch}
		dgst := f.pushPlatform(t, instance.arch, instance.revision, platform)
		headers := f.request(t, http.MethodHead, f.url+"/v2/"+f.repo+"/manifests/"+instance.arch, "", nil, http.StatusOK)
		size, err := strconv.Atoi(headers.Get("Content-Length"))
		require.NoError(t, err)
		descriptors = append(descriptors, map[string]any{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"size":      size,
			"digest":    dgst,
			"platform":  map[string]string{"os": "linux", "architecture": instance.arch},
		})
	}
	body, err := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     descriptors,
	})
	require.NoError(t, err)
	f.request(t, http.MethodPut, f.url+"/v2/"+f.repo+"/manifests/latest", "application/vnd.oci.image.index.v1+json", body, http.StatusCreated)
}

func TestRegistryPlatformCacheAcrossRestarts(t *testing.T) {
	for _, workers := range []int{1, 4} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			f := newRegistryFixture(t)
			f.pushIndex(t, "first", "first")
			target, err := url.Parse(f.url)
			require.NoError(t, err)
			proxy := httputil.NewSingleHostReverseProxy(target)
			var gets, heads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/manifests/") {
					if r.Method == http.MethodGet {
						gets.Add(1)
					} else if r.Method == http.MethodHead {
						heads.Add(1)
					}
				}
				proxy.ServeHTTP(w, r)
			}))
			defer server.Close()
			f.url = server.URL
			amd := f.image("latest")
			amd.NotifyOn = []model.NotifyOn{model.NotifyOnNew, model.NotifyOnUpdate, model.NotifyOnUnchange}
			arm := amd
			arm.Platform.Arch = "arm64"
			f.cfg.Watch.Workers = workers

			reports, _ := f.scan(t, amd)
			require.Empty(t, reports)
			reports, saved := f.scan(t, amd, arm)
			require.Len(t, reports, 2)
			require.Len(t, saved, 1, "history still has one entry per reference")
			for _, report := range reports {
				require.Equal(t, "unchange", report["status"], "adding a platform must not report a new image")
			}
			require.ElementsMatch(t, []any{"linux/amd64", "linux/arm64"}, []any{reports[0]["platform"], reports[1]["platform"]})

			for range 2 {
				beforeGets, beforeHeads := gets.Load(), heads.Load()
				reports, _ = f.scan(t, arm, amd)
				require.Len(t, reports, 2)
				require.Equal(t, beforeGets, gets.Load(), "both platforms must retain their HEAD-only fast path after restart")
				require.Equal(t, beforeHeads+2, heads.Load())
			}

			f.pushIndex(t, "second", "second")
			reports, _ = f.scan(t, amd, arm)
			require.Len(t, reports, 2)
			for _, report := range reports {
				require.Equal(t, "update", report["status"], "another platform's write must not advance this platform's baseline")
			}
			before := gets.Load()
			reports, _ = f.scan(t, arm, amd)
			require.Len(t, reports, 2)
			require.Equal(t, before, gets.Load())
			for _, report := range reports {
				require.Equal(t, "unchange", report["status"])
			}
		})
	}
}

func TestRegistryPlatformUpdatesAndMissingInstance(t *testing.T) {
	for _, compareDigest := range []bool{true, false} {
		t.Run(fmt.Sprintf("compareDigest=%t", compareDigest), func(t *testing.T) {
			f := newRegistryFixture(t)
			f.cfg.Watch.CompareDigest = new(compareDigest)
			f.cfg.Watch.Workers = 4
			amd := f.image("latest")
			amd.NotifyOn = []model.NotifyOn{model.NotifyOnNew, model.NotifyOnUpdate, model.NotifyOnUnchange}
			arm := amd
			arm.Platform.Arch = "arm64"
			f.pushIndex(t, "first", "first")
			f.scan(t, amd, arm)

			f.pushIndex(t, "second", "first")
			reports, _ := f.scan(t, arm, amd)
			require.Len(t, reports, 2)
			for _, report := range reports {
				want := "update"
				if compareDigest && report["platform"] == "linux/arm64" {
					want = "unchange"
				}
				require.Equal(t, want, report["status"])
			}

			image, err := registry.ParseImage(registry.ParseImageOptions{Name: arm.Name})
			require.NoError(t, err)
			store, err := db.New(*f.cfg.Db)
			require.NoError(t, err)
			before, err := store.GetPlatformManifest(image, arm.Platform)
			require.NoError(t, err)
			require.NoError(t, store.Close())

			f.pushIndex(t, "second", "")
			reports, _ = f.scan(t, amd, arm)
			require.Len(t, reports, 1, "an unavailable platform must not notify")
			require.Equal(t, "linux/amd64", reports[0]["platform"])
			store, err = db.New(*f.cfg.Db)
			require.NoError(t, err)
			after, err := store.GetPlatformManifest(image, arm.Platform)
			require.NoError(t, err)
			require.NoError(t, store.Close())
			require.Equal(t, before, after, "a failed fetch must preserve the platform baseline")

			f.pushIndex(t, "second", "second")
			reports, _ = f.scan(t, amd, arm)
			require.Len(t, reports, 2)
			for _, report := range reports {
				want := "update"
				if compareDigest && report["platform"] == "linux/amd64" {
					want = "unchange"
				}
				require.Equal(t, want, report["status"])
			}
		})
	}
}

func TestRegistryPlatformStateUpgrade(t *testing.T) {
	f := newRegistryFixture(t)
	f.pushIndex(t, "first", "first")
	amd := f.image("latest")
	amd.NotifyOn = []model.NotifyOn{model.NotifyOnNew, model.NotifyOnUpdate, model.NotifyOnUnchange}
	arm := amd
	arm.Platform.Arch = "arm64"
	_, saved := f.scan(t, amd)
	require.Len(t, saved, 1)

	// Use a separate database with only legacy history and version 2 metadata.
	// This tests application startup migration and the registry refresh together.
	f.cfg.Db.Path = filepath.Join(t.TempDir(), "legacy.db")
	store, err := db.New(*f.cfg.Db)
	require.NoError(t, err)
	image, err := registry.ParseImage(registry.ParseImageOptions{Name: amd.Name})
	require.NoError(t, err)
	legacy := saved[0]
	legacy.InstanceDigest = ""
	require.NoError(t, store.PutManifest(image, legacy))
	require.NoError(t, store.WriteMetadata(db.Metadata{Version: 2}))
	require.NoError(t, store.Close())

	for range 2 {
		reports, history := f.scan(t, arm, amd)
		require.Len(t, reports, 2)
		require.Len(t, history, 1)
		require.ElementsMatch(t, []any{"linux/amd64", "linux/arm64"}, []any{reports[0]["platform"], reports[1]["platform"]})
		for _, report := range reports {
			require.Equal(t, "unchange", report["status"], "upgrading must not mark existing images new or updated")
		}
	}
	store, err = db.New(*f.cfg.Db)
	require.NoError(t, err)
	defer store.Close()
	for _, watch := range []model.Image{amd, arm} {
		manifest, err := store.GetPlatformManifest(image, watch.Platform)
		require.NoError(t, err)
		require.NotEmpty(t, manifest.InstanceDigest)
		require.Equal(t, "linux/"+watch.Platform.Arch, manifest.Platform)
	}
}
