package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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
