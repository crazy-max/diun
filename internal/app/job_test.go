package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRunJobNotifiesUnchangedOnlyWhenRequested(t *testing.T) {
	remoteDigest := digest.FromString("cached manifest")

	registryServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Docker-Content-Digest", remoteDigest.String())
		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.WriteHeader(http.StatusOK)
	}))
	defer registryServer.Close()

	reports := make(chan map[string]any, 8)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		reports <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer receiver.Close()

	store, err := db.New(model.Db{Path: filepath.Join(t.TempDir(), "diun.db")})
	require.NoError(t, err)
	defer store.Close()

	sender, err := notif.New(&model.Notif{Webhook: &model.NotifWebhook{
		Endpoint: receiver.URL,
		Method:   http.MethodPost,
		Timeout:  new(time.Second),
	}}, model.Meta{})
	require.NoError(t, err)

	image, err := registry.ParseImage(registry.ParseImageOptions{Name: strings.TrimPrefix(registryServer.URL, "https://") + "/test:latest"})
	require.NoError(t, err)
	require.NoError(t, store.PutManifest(image, registry.Manifest{Name: image.Name(), Digest: remoteDigest}))

	client, err := registry.New(registry.Options{InsecureTLS: true, CompareDigest: true, Timeout: time.Second})
	require.NoError(t, err)

	di := &Diun{db: store, notif: sender, cfg: &config.Config{Watch: (&model.Watch{}).GetDefaults()}}
	job := model.Job{Provider: "file", RegImage: image, Registry: client, FirstCheck: false}

	// Defaults remain silent for unchanged images.
	job.Image.NotifyOn = model.NotifyOnDefaults
	require.Equal(t, model.ImageStatusUnchange, di.runJob(job).Status)
	require.Empty(t, reports)
	job.Image.NotifyOn = []model.NotifyOn{model.NotifyOnUnchange}
	for range 2 {
		require.Equal(t, model.ImageStatusUnchange, di.runJob(job).Status)
		select {
		case payload := <-reports:
			require.Equal(t, "unchange", payload["status"])
			require.NotContains(t, payload, "state")
		default:
			t.Fatal("unchanged scan did not notify")
		}
	}

	job.FirstCheck = true
	di.runJob(job)
	require.Empty(t, reports)

	job.FirstCheck = false
	job.Image.ExcludeTags = []string{".*"}
	require.Equal(t, model.ImageStatusSkip, di.runJob(job).Status)
	require.Empty(t, reports)
}

func TestRunJob(t *testing.T) {
	for _, tt := range []struct {
		name             string
		immutable        *bool
		cached           string
		firstCheck       bool
		firstCheckNotif  bool
		notifyOn         []model.NotifyOn
		include          []string
		exclude          []string
		registryFailure  bool
		labelURL         string
		hubLinkOverride  string
		wantStatus       model.ImageStatus
		wantUpdate       bool
		wantNotification bool
		wantHubLink      string
	}{
		{
			name:             "immutable new image still fetched",
			immutable:        new(true),
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusNew,
			wantUpdate:       true,
			wantNotification: true,
		},
		{
			name:       "immutable known image skipped even with unchanged notifications",
			immutable:  new(true),
			cached:     "stale",
			notifyOn:   []model.NotifyOn{model.NotifyOnUnchange},
			wantStatus: model.ImageStatusSkip,
		},
		{
			name:             "explicitly mutable image detects updates",
			immutable:        new(false),
			cached:           "stale",
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusUpdate,
			wantUpdate:       true,
			wantNotification: true,
		},
		{
			name:            "immutable unknown image must exist",
			immutable:       new(true),
			registryFailure: true,
			wantStatus:      model.ImageStatusError,
		},
		{
			name:             "new image",
			cached:           "",
			firstCheck:       false,
			firstCheckNotif:  false,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusNew,
			wantUpdate:       true,
			wantNotification: true,
		},
		{
			name:             "first new image",
			cached:           "",
			firstCheck:       true,
			firstCheckNotif:  false,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusNew,
			wantUpdate:       false,
			wantNotification: false,
		},
		{
			name:             "first new image notifications enabled",
			cached:           "",
			firstCheck:       true,
			firstCheckNotif:  true,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusNew,
			wantUpdate:       false,
			wantNotification: true,
		},
		{
			name:             "new image notification filtered",
			cached:           "",
			firstCheck:       false,
			firstCheckNotif:  false,
			notifyOn:         []model.NotifyOn{model.NotifyOnUpdate},
			wantStatus:       model.ImageStatusNew,
			wantUpdate:       true,
			wantNotification: false,
		},
		{
			name:             "updated image",
			cached:           "stale",
			firstCheck:       false,
			firstCheckNotif:  false,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusUpdate,
			wantUpdate:       true,
			wantNotification: true,
		},
		{
			name:             "updated image on first check",
			cached:           "stale",
			firstCheck:       true,
			firstCheckNotif:  false,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusUpdate,
			wantUpdate:       true,
			wantNotification: false,
		},
		{
			name:             "updated image first check notifications enabled",
			cached:           "stale",
			firstCheck:       true,
			firstCheckNotif:  true,
			notifyOn:         model.NotifyOnDefaults,
			wantStatus:       model.ImageStatusUpdate,
			wantUpdate:       true,
			wantNotification: true,
		},
		{
			name:             "updated image notification filtered",
			cached:           "stale",
			firstCheck:       false,
			firstCheckNotif:  false,
			notifyOn:         []model.NotifyOn{model.NotifyOnNew},
			wantStatus:       model.ImageStatusUpdate,
			wantUpdate:       true,
			wantNotification: false,
		},
		{
			name:             "unchanged first check notifications enabled",
			cached:           "matching",
			firstCheck:       true,
			firstCheckNotif:  true,
			notifyOn:         []model.NotifyOn{model.NotifyOnUnchange},
			wantStatus:       model.ImageStatusUnchange,
			wantUpdate:       false,
			wantNotification: true,
		},
		{
			name:       "include mismatch",
			cached:     "stale",
			include:    []string{"^stable$"},
			notifyOn:   model.NotifyOnDefaults,
			wantStatus: model.ImageStatusSkip,
		},
		{
			name:       "exclusion wins over inclusion",
			cached:     "stale",
			include:    []string{"^latest$"},
			exclude:    []string{"^latest$"},
			notifyOn:   model.NotifyOnDefaults,
			wantStatus: model.ImageStatusSkip,
		},
		{
			name:            "registry failure",
			cached:          "stale",
			registryFailure: true,
			notifyOn:        model.NotifyOnDefaults,
			wantStatus:      model.ImageStatusError,
		},
		{
			name:        "default hub link",
			cached:      "matching",
			wantStatus:  model.ImageStatusUnchange,
			wantHubLink: "https://default.example/test",
		},
		{
			name:        "OCI label overrides default hub link",
			labelURL:    "https://label.example/test",
			wantStatus:  model.ImageStatusNew,
			wantUpdate:  true,
			wantHubLink: "https://label.example/test",
		},
		{
			name:            "explicit hub link overrides OCI label",
			labelURL:        "https://label.example/test",
			hubLinkOverride: "https://override.example/test",
			wantStatus:      model.ImageStatusNew,
			wantUpdate:      true,
			wantHubLink:     "https://override.example/test",
		},
		{
			name:            "explicit hub link overrides default",
			hubLinkOverride: "https://override.example/test",
			wantStatus:      model.ImageStatusNew,
			wantUpdate:      true,
			wantHubLink:     "https://override.example/test",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{}
			if tt.labelURL != "" {
				labels["org.opencontainers.image.url"] = tt.labelURL
			}

			configBlob, err := json.Marshal(map[string]any{
				"architecture": "amd64",
				"os":           "linux",
				"config":       map[string]any{"Labels": labels},
				"rootfs":       map[string]any{"type": "layers", "diff_ids": []string{}},
			})
			require.NoError(t, err)
			configDigest := digest.FromBytes(configBlob)

			manifestBlob, err := json.Marshal(map[string]any{
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

			remoteDigest := digest.FromBytes(manifestBlob)
			var requests atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				switch r.URL.Path {
				case "/v2/":
					w.WriteHeader(http.StatusOK)
				case "/v2/test/manifests/latest":
					if tt.registryFailure {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
					w.Header().Set("Docker-Content-Digest", remoteDigest.String())
					if r.Method != http.MethodHead {
						_, _ = w.Write(manifestBlob)
					}
				case "/v2/test/blobs/" + configDigest.String():
					_, _ = w.Write(configBlob)
				default:
					t.Errorf("unexpected registry request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			reports := make(chan map[string]any, 8)
			receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload map[string]any
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				reports <- payload
				w.WriteHeader(http.StatusOK)
			}))
			defer receiver.Close()

			store, err := db.New(model.Db{Path: filepath.Join(t.TempDir(), "diun.db")})
			require.NoError(t, err)
			defer store.Close()

			sender, err := notif.New(&model.Notif{Webhook: &model.NotifWebhook{
				Endpoint: receiver.URL,
				Method:   http.MethodPost,
				Timeout:  new(5 * time.Second),
			}}, model.Meta{})
			require.NoError(t, err)

			image, err := registry.ParseImage(registry.ParseImageOptions{
				Name: strings.TrimPrefix(server.URL, "https://") + "/test:latest",
			})
			require.NoError(t, err)
			image.HubLink = "https://default.example/test"

			var cached registry.Manifest
			if tt.cached != "" {
				cached = registry.Manifest{Name: image.Name(), Tag: image.Tag, Digest: digest.FromString("old manifest")}
				cached.Labels = map[string]string{"example": "preserved"}
				cached.Platform = "linux/amd64"
				cached.Created = new(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
				if tt.cached == "matching" {
					cached.Digest = remoteDigest
				}
				require.NoError(t, store.PutManifest(image, cached))
			}

			client, err := registry.New(registry.Options{
				InsecureTLS:   true,
				CompareDigest: true,
				Timeout:       5 * time.Second,
			})
			require.NoError(t, err)

			watch := (&model.Watch{}).GetDefaults()
			watch.FirstCheckNotif = new(tt.firstCheckNotif)

			di := &Diun{db: store, notif: sender, cfg: &config.Config{Watch: watch}}
			entry := di.runJob(model.Job{
				Provider:        "file",
				RegImage:        image,
				Registry:        client,
				FirstCheck:      tt.firstCheck,
				HubLinkOverride: tt.hubLinkOverride,
				Image: model.Image{
					NotifyOn:            tt.notifyOn,
					IncludeTags:         tt.include,
					ExcludeTags:         tt.exclude,
					AssumeImmutableTags: tt.immutable,
				},
			})
			require.Equal(t, tt.wantStatus, entry.Status)
			require.Equal(t, tt.wantUpdate, entry.UpdateAvailable())
			if tt.wantHubLink != "" {
				require.Equal(t, tt.wantHubLink, entry.Image.HubLink)
			}

			saved, err := store.GetManifest(image)
			require.NoError(t, err)
			if tt.wantStatus == model.ImageStatusSkip || tt.wantStatus == model.ImageStatusError {
				require.Equal(t, cached, saved)
			} else {
				require.Equal(t, remoteDigest, saved.Digest)
				require.Equal(t, entry.Manifest, saved)
			}
			if tt.wantStatus == model.ImageStatusSkip {
				require.Zero(t, requests.Load(), "skipped images must not contact the registry")
			}
			if tt.immutable != nil && *tt.immutable && tt.cached != "" {
				require.Equal(t, cached, entry.Manifest)
				// Disabling the assumption resumes comparison against the preserved digest.
				resumed := di.runJob(model.Job{
					Provider: "file",
					RegImage: image,
					Registry: client,
					Image:    model.Image{AssumeImmutableTags: new(false)},
				})
				require.Equal(t, model.ImageStatusUpdate, resumed.Status)
				require.True(t, resumed.UpdateAvailable())
				require.Equal(t, remoteDigest, resumed.Manifest.Digest)
			}
			if tt.wantNotification {
				require.Len(t, reports, 1)
				payload := <-reports
				require.Equal(t, string(tt.wantStatus), payload["status"])
				require.Equal(t, entry.Image.String(), payload["image"])
				require.Equal(t, remoteDigest.String(), payload["digest"])
			}
			require.Empty(t, reports)
		})
	}
}
