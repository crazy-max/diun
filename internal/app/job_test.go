package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/registry/auth"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

// Run a single consumer through the same group execution as a scheduled scan.
func (di *Diun) runJob(job model.Job) model.NotifEntry {
	baseline, err := di.db.GetPlatformManifest(job.RegImage, job.Image.Platform)
	return di.runJobs(&analysisJob{consumers: []model.Job{job}, baseline: baseline, err: err})[0]
}

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

func TestSharedAnalysis(t *testing.T) {
	for _, tt := range []struct {
		name         string
		seed         bool
		immutable    bool
		excluded     bool
		failure      bool
		artifact     bool
		writeFailure bool
		firstNotif   bool
		want         model.ImageStatus
		wantReports  int
	}{
		{
			name:        "updated consumers share the original baseline",
			seed:        true,
			want:        model.ImageStatusUpdate,
			wantReports: 2,
		},
		{
			name: "first scan suppresses every consumer",
			want: model.ImageStatusNew,
		},
		{
			name:        "first scan notifications preserve every consumer",
			firstNotif:  true,
			want:        model.ImageStatusNew,
			wantReports: 2,
		},
		{
			name:        "immutable consumer does not inherit mutable update",
			seed:        true,
			immutable:   true,
			want:        model.ImageStatusUpdate,
			wantReports: 1,
		},
		{
			name:        "immutable consumer without baseline shares discovery",
			immutable:   true,
			firstNotif:  true,
			want:        model.ImageStatusNew,
			wantReports: 2,
		},
		{
			name:        "excluded consumer does not suppress another watch",
			seed:        true,
			excluded:    true,
			want:        model.ImageStatusUpdate,
			wantReports: 1,
		},
		{
			name:    "failure reaches every active consumer",
			seed:    true,
			failure: true,
			want:    model.ImageStatusError,
		},
		{
			name:     "non-image artifacts skip every active consumer",
			seed:     true,
			artifact: true,
			want:     model.ImageStatusSkip,
		},
		{
			name:         "database write failure preserves analysis but suppresses notification",
			seed:         true,
			writeFailure: true,
			want:         model.ImageStatusUpdate,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			configBlob := []byte(`{"architecture":"amd64","os":"linux","config":{"Labels":{"org.opencontainers.image.url":"https://image.example"}},"rootfs":{"type":"layers","diff_ids":[]}}`)
			configDigest := digest.FromBytes(configBlob)
			configType := "application/vnd.oci.image.config.v1+json"
			if tt.artifact {
				configType = "application/vnd.example.artifact.config.v1+json"
			}
			manifestBlob, err := json.Marshal(map[string]any{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.manifest.v1+json",
				"config": map[string]any{
					"mediaType": configType,
					"digest":    configDigest,
					"size":      len(configBlob),
				},
				"layers": []any{},
			})
			require.NoError(t, err)
			remoteDigest := digest.FromBytes(manifestBlob)
			var heads atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/":
					w.WriteHeader(http.StatusOK)
				case "/v2/test/manifests/latest":
					if r.Method == http.MethodHead {
						heads.Add(1)
					}
					if tt.failure {
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
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
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
			watch := (&model.Watch{}).GetDefaults()
			watch.FirstCheckNotif = new(tt.firstNotif)
			di := &Diun{db: store, notif: sender, cfg: &config.Config{
				Watch: watch,
				RegOpts: model.RegOpts{
					{
						Name:        "local",
						Selector:    model.RegOptSelectorName,
						InsecureTLS: new(true),
						Timeout:     new(time.Second),
					},
				},
			}}
			name := strings.TrimPrefix(server.URL, "https://") + "/test"
			image, err := registry.ParseImage(registry.ParseImageOptions{Name: name})
			require.NoError(t, err)
			old := registry.Manifest{Name: image.Name(), Tag: "latest", Digest: digest.FromString("old")}
			if tt.seed {
				require.NoError(t, store.PutManifest(image, old))
			}
			jobs := []model.Job{
				{
					Provider: "docker",
					Image: model.Image{
						Name:      name,
						RegOpt:    "local",
						WatchMode: registry.WatchModeTag,
						NotifyOn:  model.NotifyOnDefaults,
						Metadata:  map[string]string{"service": "first"},
						HubLink:   "https://first.example",
					},
				},
				{
					Provider: "file",
					Image: model.Image{
						Name:      name + ":latest",
						RegOpt:    "local",
						WatchMode: registry.WatchModeTag,
						NotifyOn:  model.NotifyOnDefaults,
						Metadata:  map[string]string{"service": "second"},
						HubLink:   "https://second.example",
					},
				},
			}
			jobs[0].Image.AssumeImmutableTags = new(tt.immutable)
			if tt.excluded {
				jobs[0].Image.ExcludeTags = []string{".*"}
			}
			var plan jobPlan
			for _, job := range jobs {
				di.createJob(job, &plan)
			}
			require.Len(t, plan.jobs, 1)
			require.Zero(t, heads.Load(), "discovery must finish before analysis")
			if tt.writeFailure {
				require.NoError(t, store.Close())
			}
			entries := di.runJobs(plan.jobs[0])
			require.Len(t, entries, 2)
			require.Equal(t, int32(1), heads.Load())
			require.Equal(t, tt.want, entries[1].Status)
			if tt.excluded || (tt.immutable && tt.seed) {
				require.Equal(t, model.ImageStatusSkip, entries[0].Status)
			} else {
				require.Equal(t, tt.want, entries[0].Status)
			}
			require.Len(t, reports, tt.wantReports)
			if tt.writeFailure {
				for _, entry := range entries {
					require.Equal(t, tt.want, entry.Status)
					require.Equal(t, remoteDigest, entry.Manifest.Digest)
					require.True(t, entry.UpdateAvailable())
				}
				return
			}
			for range tt.wantReports {
				payload := <-reports
				service := payload["metadata"].(map[string]any)["service"].(string)
				require.Equal(t, "https://"+service+".example", payload["hub_link"])
				if service == "first" {
					require.Equal(t, "docker", payload["provider"])
				} else {
					require.Equal(t, "file", payload["provider"])
				}
			}
			saved, err := store.GetManifest(image)
			require.NoError(t, err)
			if tt.failure || tt.artifact {
				require.Equal(t, old, saved)
			} else {
				require.Equal(t, remoteDigest, saved.Digest)
			}
			// A new plan must perform a new check, and each consumer makes its own
			// notify-on decision even though the analysis is shared.
			jobs[0].Image.NotifyOn = []model.NotifyOn{model.NotifyOnUnchange}
			var next jobPlan
			for _, job := range jobs {
				di.createJob(job, &next)
			}
			nextEntries := di.runJobs(next.jobs[0])
			require.Equal(t, int32(2), heads.Load())
			if tt.artifact {
				require.Equal(t, model.ImageStatusSkip, nextEntries[0].Status)
				require.Equal(t, model.ImageStatusSkip, nextEntries[1].Status)
				require.Empty(t, reports)
			} else if !tt.failure {
				require.Equal(t, model.ImageStatusUnchange, nextEntries[1].Status)
				if !tt.immutable && !tt.excluded {
					require.Len(t, reports, 1)
				} else {
					require.Empty(t, reports)
				}
			}
		})
	}
}

func TestSharedAnalysisRegistryOptions(t *testing.T) {
	for _, tt := range []struct {
		name       string
		change     func(*model.RegOpt, *model.Image)
		wantGroups int
	}{
		{
			name:       "different names with identical effective options",
			change:     func(_ *model.RegOpt, _ *model.Image) {},
			wantGroups: 1,
		},
		{
			name:       "different password",
			change:     func(r *model.RegOpt, _ *model.Image) { r.Password = "other" },
			wantGroups: 2,
		},
		{
			name: "nested static credentials match legacy credentials",
			change: func(r *model.RegOpt, _ *model.Image) {
				r.Auth = &model.RegistryAuth{Static: &auth.StaticOptions{Username: r.Username, Password: r.Password}}
				r.Username, r.Password = "", ""
			},
			wantGroups: 1,
		},
		{
			name:       "different TLS policy",
			change:     func(r *model.RegOpt, _ *model.Image) { r.InsecureTLS = new(true) },
			wantGroups: 2,
		},
		{
			name:       "different timeout",
			change:     func(r *model.RegOpt, _ *model.Image) { r.Timeout = new(time.Second) },
			wantGroups: 2,
		},
		{
			name:       "different platform",
			change:     func(_ *model.RegOpt, i *model.Image) { i.Platform.Arch = "arm64" },
			wantGroups: 2,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			di := newTestDiun(t, "")
			defer di.db.Close()
			first := (&model.RegOpt{}).GetDefaults()
			first.Name, first.Username, first.Password = "first", "user", "password"
			second := *first
			second.Name = "second"
			image := model.Image{
				Name:      "example.com/test:latest",
				WatchMode: registry.WatchModeTag,
				RegOpt:    first.Name,
				Platform:  model.ImagePlatform{OS: "linux", Arch: "amd64"},
			}
			other := image
			other.RegOpt = second.Name
			tt.change(&second, &other)
			di.cfg.RegOpts = model.RegOpts{*first, second}
			var plan jobPlan
			di.createJob(model.Job{Provider: "file", Image: image}, &plan)
			di.createJob(model.Job{Provider: "docker", Image: other}, &plan)
			require.Len(t, plan.jobs, tt.wantGroups)
		})
	}
}

func TestSharedAnalysisTagExpansion(t *testing.T) {
	var lists atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
		case "/v2/test/tags/list":
			lists.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"test","tags":["1.0.0","2.0.0","3.0.0"]}`))
		default:
			t.Errorf("analysis started during discovery: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	for _, reverse := range []bool{false, true} {
		di := newTestDiun(t, "")
		di.cfg.RegOpts = model.RegOpts{
			{
				Name:        "local",
				Selector:    model.RegOptSelectorName,
				InsecureTLS: new(true),
				Timeout:     new(time.Second),
			},
		}
		name := strings.TrimPrefix(server.URL, "https://") + "/test:"
		images := []model.Image{
			{
				Name:      name + "1.0.0",
				RegOpt:    "local",
				WatchMode: registry.WatchModeAll,
				SortTags:  registry.SortTagReverse,
				MaxTags:   1,
			},
			{
				Name:      name + "1.0.0",
				RegOpt:    "local",
				WatchMode: registry.WatchModeSemver,
				SortTags:  registry.SortTagLexicographical,
				MaxTags:   1,
			},
			{
				Name:      name + "3.0.0",
				RegOpt:    "local",
				WatchMode: registry.WatchModeTag,
			},
		}
		if reverse {
			slices.Reverse(images)
		}
		var plan jobPlan
		before := lists.Load()
		for _, image := range images {
			di.createJob(model.Job{Provider: "file", Image: image}, &plan)
		}
		require.Equal(t, before+1, lists.Load())
		require.Len(t, plan.jobs, 3)
		counts := map[string]int{}
		for _, group := range plan.jobs {
			counts[group.consumers[0].RegImage.Tag] = len(group.consumers)
			for _, consumer := range group.consumers {
				require.True(t, consumer.FirstCheck)
			}
		}
		require.Equal(t, map[string]int{"1.0.0": 2, "2.0.0": 1, "3.0.0": 2}, counts)
		require.NoError(t, di.db.Close())
	}
}

func TestSharedAnalysisUnknownRegistryOptions(t *testing.T) {
	di := newTestDiun(t, "")
	defer di.db.Close()
	var plan jobPlan
	di.createJob(model.Job{Image: model.Image{
		Name:      "example.com/test:latest",
		RegOpt:    "missing",
		WatchMode: registry.WatchModeTag,
	}}, &plan)
	require.Empty(t, plan.jobs)
}

func TestSharedAnalysisTagListFailureRecovery(t *testing.T) {
	var lists atomic.Int32
	var fail atomic.Bool
	fail.Store(true)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != "/v2/test/tags/list" {
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		lists.Add(1)
		if fail.Load() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"test","tags":["latest","1.0.0"]}`))
	}))
	defer server.Close()
	client, err := registry.New(registry.Options{InsecureTLS: true, Timeout: time.Second})
	require.NoError(t, err)
	image, err := registry.ParseImage(registry.ParseImageOptions{Name: strings.TrimPrefix(server.URL, "https://") + "/test:latest"})
	require.NoError(t, err)
	job := model.Job{RegImage: image, Registry: client}
	opts := registry.TagsOptions{Image: image, WatchMode: registry.WatchModeAll}
	var plan jobPlan
	_, err = plan.selectTags(job, opts)
	require.Error(t, err)
	fail.Store(false)
	_, err = plan.selectTags(job, opts)
	require.Error(t, err, "later consumers share the failed discovery in this scan")
	require.Equal(t, int32(1), lists.Load())
	var next jobPlan
	tags, err := next.selectTags(job, opts)
	require.NoError(t, err)
	require.Equal(t, []string{"latest", "1.0.0"}, tags.List)
	require.Equal(t, int32(2), lists.Load(), "the next scan must retry tag discovery")
}

func TestSharedAnalysisDatabaseReadFailure(t *testing.T) {
	di := newTestDiun(t, "")
	require.NoError(t, di.db.Close())
	image, err := registry.ParseImage(registry.ParseImageOptions{Name: "example.com/test:latest"})
	require.NoError(t, err)
	baseline, err := di.db.GetManifest(image)
	require.Error(t, err)
	entries := di.runJobs(&analysisJob{
		baseline: baseline,
		err:      err,
		consumers: []model.Job{
			{
				RegImage: image,
			},
			{
				RegImage: image,
				Image:    model.Image{ExcludeTags: []string{".*"}},
			},
		},
	})
	require.Equal(t, model.ImageStatusError, entries[0].Status)
	require.Equal(t, model.ImageStatusSkip, entries[1].Status)
}

func TestLoadAuthFailures(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{"), 0o600))
	missing := filepath.Join(dir, "missing")
	for _, tc := range []struct {
		name string
		reg  model.RegOpt
		want auth.Provider
	}{
		{
			name: "explicit static failure stops",
			reg: model.RegOpt{
				Auth: &model.RegistryAuth{Static: &auth.StaticOptions{Username: "user", PasswordFile: missing}},
			},
		},
		{
			name: "legacy password failure stops",
			reg: model.RegOpt{
				Username:     "user",
				PasswordFile: missing,
			},
		},
		{
			name: "legacy username failure stops",
			reg: model.RegOpt{
				UsernameFile: missing,
			},
		},
		{
			name: "Docker discovery failure continues anonymously",
			want: auth.Docker{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var plan jobPlan
			provider, err := plan.loadAuth(&tc.reg, "diun-auth-test.invalid")
			require.Error(t, err)
			require.Equal(t, tc.want, provider)
		})
	}
}
