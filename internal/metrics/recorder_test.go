package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	regpkg "github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestRecorderRecordRun(t *testing.T) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder("1.2.3")
	registry.MustRegister(recorder)

	alpine, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:3.19"})
	require.NoError(t, err)
	busybox, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "busybox:latest"})
	require.NoError(t, err)

	created := time.Unix(1000, 0).UTC()
	completedAt := time.Unix(2000, 0).UTC()
	updateEntry := model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "docker",
		Image:    alpine,
		Manifest: regpkg.Manifest{Created: &created},
	}
	updateEntry.MarkUpdateAvailable()
	recorder.RecordRun(&model.NotifEntries{
		Entries: []model.NotifEntry{
			updateEntry,
			{
				Status:   model.ImageStatusNew,
				Provider: "file",
				Image:    busybox,
			},
		},
	}, 1500*time.Millisecond, completedAt)
	recorder.RecordSkippedRun()

	problems, err := testutil.GatherAndLint(registry,
		"diun_build_info",
		"diun_image_created_timestamp_seconds",
		"diun_image_last_check_status",
		"diun_image_last_check_timestamp_seconds",
		"diun_image_update_available",
		"diun_watch_last_run_duration_seconds",
		"diun_watch_last_run_images",
		"diun_watch_last_run_timestamp_seconds",
		"diun_watch_runs_total",
		"diun_watch_skipped_runs_total",
	)
	require.NoError(t, err)
	require.Empty(t, problems)

	err = testutil.GatherAndCompare(registry, strings.NewReader(`
# HELP diun_build_info Build information for the Diun instance.
# TYPE diun_build_info gauge
diun_build_info{version="1.2.3"} 1
# HELP diun_image_created_timestamp_seconds Unix timestamp of the image manifest creation time reported by the registry.
# TYPE diun_image_created_timestamp_seconds gauge
diun_image_created_timestamp_seconds{image="docker.io/library/alpine:3.19",provider="docker"} 1000
# HELP diun_image_last_check_status Last check status for the image. The active status has value 1.
# TYPE diun_image_last_check_status gauge
diun_image_last_check_status{image="docker.io/library/alpine:3.19",provider="docker",status="error"} 0
diun_image_last_check_status{image="docker.io/library/alpine:3.19",provider="docker",status="new"} 0
diun_image_last_check_status{image="docker.io/library/alpine:3.19",provider="docker",status="skip"} 0
diun_image_last_check_status{image="docker.io/library/alpine:3.19",provider="docker",status="unchange"} 0
diun_image_last_check_status{image="docker.io/library/alpine:3.19",provider="docker",status="update"} 1
diun_image_last_check_status{image="docker.io/library/busybox:latest",provider="file",status="error"} 0
diun_image_last_check_status{image="docker.io/library/busybox:latest",provider="file",status="new"} 1
diun_image_last_check_status{image="docker.io/library/busybox:latest",provider="file",status="skip"} 0
diun_image_last_check_status{image="docker.io/library/busybox:latest",provider="file",status="unchange"} 0
diun_image_last_check_status{image="docker.io/library/busybox:latest",provider="file",status="update"} 0
# HELP diun_image_last_check_timestamp_seconds Unix timestamp of the last completed check for the image.
# TYPE diun_image_last_check_timestamp_seconds gauge
diun_image_last_check_timestamp_seconds{image="docker.io/library/alpine:3.19",provider="docker"} 2000
diun_image_last_check_timestamp_seconds{image="docker.io/library/busybox:latest",provider="file"} 2000
# HELP diun_image_update_available Whether the last check found an actionable update for the image.
# TYPE diun_image_update_available gauge
diun_image_update_available{image="docker.io/library/alpine:3.19",provider="docker"} 1
diun_image_update_available{image="docker.io/library/busybox:latest",provider="file"} 0
# HELP diun_watch_last_run_duration_seconds Duration in seconds of the last completed Diun watch run.
# TYPE diun_watch_last_run_duration_seconds gauge
diun_watch_last_run_duration_seconds 1.5
# HELP diun_watch_last_run_images Number of images by status in the last completed Diun watch run.
# TYPE diun_watch_last_run_images gauge
diun_watch_last_run_images{status="error"} 0
diun_watch_last_run_images{status="new"} 1
diun_watch_last_run_images{status="skip"} 0
diun_watch_last_run_images{status="unchange"} 0
diun_watch_last_run_images{status="update"} 1
# HELP diun_watch_last_run_timestamp_seconds Unix timestamp of the last completed Diun watch run.
# TYPE diun_watch_last_run_timestamp_seconds gauge
diun_watch_last_run_timestamp_seconds 2000
# HELP diun_watch_runs_total Total number of completed Diun watch runs.
# TYPE diun_watch_runs_total counter
diun_watch_runs_total 1
# HELP diun_watch_skipped_runs_total Total number of Diun watch runs skipped because another run was already active.
# TYPE diun_watch_skipped_runs_total counter
diun_watch_skipped_runs_total 1
`),
		"diun_build_info",
		"diun_image_created_timestamp_seconds",
		"diun_image_last_check_status",
		"diun_image_last_check_timestamp_seconds",
		"diun_image_update_available",
		"diun_watch_last_run_duration_seconds",
		"diun_watch_last_run_images",
		"diun_watch_last_run_timestamp_seconds",
		"diun_watch_runs_total",
		"diun_watch_skipped_runs_total",
	)
	require.NoError(t, err)
}

// fakeLocalDigestSource resolves container IDs against a fixed table, and
// leaves unknown IDs out of the result - the same contract a real
// LocalDigestSource has for a container that's since been removed.
type fakeLocalDigestSource map[string]digest.Digest

func (f fakeLocalDigestSource) LocalDigests(containerIDs []string) map[string]digest.Digest {
	result := make(map[string]digest.Digest)
	for _, id := range containerIDs {
		if dgst, ok := f[id]; ok {
			result[id] = dgst
		}
	}
	return result
}

func TestRecorderLocalOutdated(t *testing.T) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder("1.2.3")
	registry.MustRegister(recorder)

	alpine, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:3.19"})
	require.NoError(t, err)
	nginx, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "nginx:1.27"})
	require.NoError(t, err)
	busybox, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "busybox:latest"})
	require.NoError(t, err)

	recorder.RegisterLocalDigestSource("docker", fakeLocalDigestSource{
		// ctn-alpine is still running the digest Diun last recorded.
		"ctn-alpine": "sha256:aaaa000000000000000000000000000000000000000000000000000000aaaa",
		// ctn-nginx was pulled + recreated since Diun's last registry poll.
		"ctn-nginx": "sha256:bbbb000000000000000000000000000000000000000000000000000000bbbb",
		// ctn-busybox is deliberately absent, simulating a container that
		// was removed between the watch run and this scrape.
	})

	recorder.RecordRun(&model.NotifEntries{
		Entries: []model.NotifEntry{
			{
				Status:   model.ImageStatusUnchange,
				Provider: "docker",
				Image:    alpine,
				Manifest: regpkg.Manifest{Digest: "sha256:aaaa000000000000000000000000000000000000000000000000000000aaaa"},
				Metadata: map[string]string{"ctn_id": "ctn-alpine"},
			},
			{
				Status:   model.ImageStatusUnchange,
				Provider: "docker",
				Image:    nginx,
				Manifest: regpkg.Manifest{Digest: "sha256:cccc000000000000000000000000000000000000000000000000000000cccc"},
				Metadata: map[string]string{"ctn_id": "ctn-nginx"},
			},
			{
				Status:   model.ImageStatusUnchange,
				Provider: "docker",
				Image:    busybox,
				Manifest: regpkg.Manifest{Digest: "sha256:dddd000000000000000000000000000000000000000000000000000000dddd"},
				Metadata: map[string]string{"ctn_id": "ctn-busybox"},
			},
			{
				// No provider is registered for "file", so this must be
				// left out of diun_image_local_outdated entirely, not
				// reported as 0.
				Status:   model.ImageStatusUnchange,
				Provider: "file",
				Image:    busybox,
				Manifest: regpkg.Manifest{Digest: "sha256:dddd000000000000000000000000000000000000000000000000000000dddd"},
			},
		},
	}, time.Second, time.Unix(2000, 0).UTC())

	err = testutil.GatherAndCompare(registry, strings.NewReader(`
# HELP diun_image_local_outdated Whether the container currently running for the image no longer matches the digest Diun last recorded from the registry. Recomputed from the local Docker daemon on every scrape, independent of the watch schedule. Omitted when the local digest can't be resolved.
# TYPE diun_image_local_outdated gauge
diun_image_local_outdated{image="docker.io/library/alpine:3.19",provider="docker"} 0
diun_image_local_outdated{image="docker.io/library/nginx:1.27",provider="docker"} 1
`), "diun_image_local_outdated")
	require.NoError(t, err)
}
