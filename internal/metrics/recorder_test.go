package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	regpkg "github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalDigestsProviders(t *testing.T) {
	r, reg := NewRecorder("test", nil)
	index, platform, old := digest.FromString("index"), digest.FromString("platform"), digest.FromString("old")
	img, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:latest@" + index.String()})
	require.NoError(t, err)
	raw, err := json.Marshal(ocispecs.Index{Manifests: []ocispecs.Descriptor{{Digest: platform}}})
	require.NoError(t, err)
	entries := &model.NotifEntries{}
	for _, name := range []string{"docker", "swarm", "kubernetes", "file", "unavailable"} {
		entries.Add(model.NotifEntry{Provider: name, Image: img, Status: model.ImageStatusUnchange, Manifest: regpkg.Manifest{Digest: index, Raw: raw}})
	}
	r.RecordRun(entries, time.Second, time.Now())
	started := make(chan struct{})
	r.SetLocalDigests("unavailable", func(ctx context.Context) (map[string][][]digest.Digest, error) {
		// A failed provider must not suppress the other providers.
		select {
		case <-started:
		case <-ctx.Done():
		}
		return nil, context.DeadlineExceeded
	})
	r.SetLocalDigests("docker", func(context.Context) (map[string][][]digest.Digest, error) {
		return map[string][][]digest.Digest{"docker.io/library/alpine:latest": {{platform}}}, nil
	})
	r.SetLocalDigests("swarm", func(context.Context) (map[string][][]digest.Digest, error) {
		return map[string][][]digest.Digest{"docker.io/library/alpine:latest": {{index}, {old}}}, nil
	})
	r.SetLocalDigests("kubernetes", func(context.Context) (map[string][][]digest.Digest, error) {
		close(started)
		return map[string][][]digest.Digest{"docker.io/library/alpine:latest": {{platform}}}, nil
	})
	err = testutil.GatherAndCompare(reg, strings.NewReader(fmt.Sprintf(`
# HELP diun_image_local_digest_mismatch Whether a watched container differs from the last recorded registry digest.
# TYPE diun_image_local_digest_mismatch gauge
diun_image_local_digest_mismatch{image="%s",provider="docker"} 1
diun_image_local_digest_mismatch{image="%s",provider="kubernetes"} 0
diun_image_local_digest_mismatch{image="%s",provider="swarm"} 1
`, img.String(), img.String(), img.String())), "diun_image_local_digest_mismatch")
	require.NoError(t, err)
}

func TestRecorderRecordRun(t *testing.T) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder("1.2.3", nil)
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

func TestRecorderMetadataLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder("1.2.3", []string{"channel", "environment"})
	registry.MustRegister(recorder)

	alpine, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:3.19"})
	require.NoError(t, err)
	busybox, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "busybox:latest"})
	require.NoError(t, err)

	created := time.Unix(1000, 0).UTC()
	metadata := map[string]string{
		"channel": "newest",
		"foo":     "bar",
	}
	recorder.RecordRun(&model.NotifEntries{
		Entries: []model.NotifEntry{
			{
				Status:   model.ImageStatusUpdate,
				Provider: "docker",
				Image:    alpine,
				Manifest: regpkg.Manifest{Created: &created},
				Metadata: metadata,
			},
			{
				Status:   model.ImageStatusNew,
				Provider: "file",
				Image:    busybox,
				Manifest: regpkg.Manifest{Created: &created},
				Metadata: map[string]string{
					"environment": "production",
				},
			},
		},
	}, time.Second, time.Unix(2000, 0).UTC())

	// the recorder must not observe later writes to provider-owned metadata
	metadata["channel"] = "mutated"

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	imageMetricFamilies := map[string]bool{
		"diun_image_created_timestamp_seconds":    false,
		"diun_image_last_check_status":            false,
		"diun_image_last_check_timestamp_seconds": false,
		"diun_image_update_available":             false,
	}
	for _, family := range metricFamilies {
		if _, ok := imageMetricFamilies[family.GetName()]; !ok {
			continue
		}
		imageMetricFamilies[family.GetName()] = true
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			require.NotContains(t, labels, "metadata_foo")
			require.Contains(t, labels, "metadata_channel")
			require.Contains(t, labels, "metadata_environment")
			switch labels["image"] {
			case "docker.io/library/alpine:3.19":
				require.Equal(t, "newest", labels["metadata_channel"])
				require.Equal(t, "", labels["metadata_environment"])
			case "docker.io/library/busybox:latest":
				require.Equal(t, "", labels["metadata_channel"])
				require.Equal(t, "production", labels["metadata_environment"])
			default:
				t.Fatalf("unexpected image label %q", labels["image"])
			}
		}
	}
	for family, found := range imageMetricFamilies {
		require.True(t, found, "metric family %s was not gathered", family)
	}
}

func TestRecorderMetadataSeriesIdentity(t *testing.T) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder("1.2.3", []string{"ctn_id"})
	registry.MustRegister(recorder)

	alpine, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:3.19"})
	require.NoError(t, err)

	// Two containers running the same image, told apart only by their metadata.
	current := digest.FromString("current")
	recorder.SetLocalDigests("docker", func(context.Context) (map[string][][]digest.Digest, error) {
		return map[string][][]digest.Digest{alpine.String(): {{current}, {digest.FromString("old")}}}, nil
	})
	first := model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "docker",
		Image:    alpine,
		Metadata: map[string]string{"ctn_id": "aaa"},
		Manifest: regpkg.Manifest{Digest: current},
	}
	first.MarkUpdateAvailable()
	second := model.NotifEntry{
		Status:   model.ImageStatusUnchange,
		Provider: "docker",
		Image:    alpine,
		Metadata: map[string]string{"ctn_id": "bbb"},
		Manifest: regpkg.Manifest{Digest: current},
	}
	recorder.RecordRun(&model.NotifEntries{
		Entries: []model.NotifEntry{first, second},
	}, time.Second, time.Unix(2000, 0).UTC())

	err = testutil.GatherAndCompare(registry, strings.NewReader(`
# HELP diun_image_last_check_timestamp_seconds Unix timestamp of the last completed check for the image.
# TYPE diun_image_last_check_timestamp_seconds gauge
diun_image_last_check_timestamp_seconds{image="docker.io/library/alpine:3.19",metadata_ctn_id="aaa",provider="docker"} 2000
diun_image_last_check_timestamp_seconds{image="docker.io/library/alpine:3.19",metadata_ctn_id="bbb",provider="docker"} 2000
# HELP diun_image_local_digest_mismatch Whether a watched container differs from the last recorded registry digest.
# TYPE diun_image_local_digest_mismatch gauge
diun_image_local_digest_mismatch{image="docker.io/library/alpine:3.19",metadata_ctn_id="aaa",provider="docker"} 1
diun_image_local_digest_mismatch{image="docker.io/library/alpine:3.19",metadata_ctn_id="bbb",provider="docker"} 1
# HELP diun_image_update_available Whether the last check found an actionable update for the image.
# TYPE diun_image_update_available gauge
diun_image_update_available{image="docker.io/library/alpine:3.19",metadata_ctn_id="aaa",provider="docker"} 1
diun_image_update_available{image="docker.io/library/alpine:3.19",metadata_ctn_id="bbb",provider="docker"} 0
`),
		"diun_image_last_check_timestamp_seconds",
		"diun_image_update_available",
		"diun_image_local_digest_mismatch",
	)
	require.NoError(t, err)
}

func TestLocalDigestMismatch(t *testing.T) {
	r, reg := NewRecorder("test", nil)
	img, err := regpkg.ParseImage(regpkg.ParseImageOptions{Name: "alpine:latest"})
	require.NoError(t, err)
	old, current := digest.FromString("old"), digest.FromString("current")
	local := [][]digest.Digest{{old}}
	r.SetLocalDigests("docker", func(ctx context.Context) (map[string][][]digest.Digest, error) {
		_, bounded := ctx.Deadline()
		assert.True(t, bounded)
		// A scrape must not hold the recorder lock while doing I/O.
		r.RecordSkippedRun()
		return map[string][][]digest.Digest{img.String(): local}, nil
	})
	entry := model.NotifEntry{Provider: "docker", Image: img, Status: model.ImageStatusUpdate, Manifest: regpkg.Manifest{Digest: current}}
	entry.MarkUpdateAvailable()
	check := func(want int) {
		t.Helper()
		err := testutil.GatherAndCompare(reg, strings.NewReader(fmt.Sprintf(`
# HELP diun_image_local_digest_mismatch Whether a watched container differs from the last recorded registry digest.
# TYPE diun_image_local_digest_mismatch gauge
diun_image_local_digest_mismatch{image="docker.io/library/alpine:latest",provider="docker"} %d
`, want)), "diun_image_local_digest_mismatch")
		require.NoError(t, err)
	}
	r.RecordRun(&model.NotifEntries{Entries: []model.NotifEntry{entry}}, time.Second, time.Now())
	check(1)
	entry = model.NotifEntry{Provider: "docker", Image: img, Status: model.ImageStatusUnchange, Manifest: regpkg.Manifest{Digest: current}}
	r.RecordRun(&model.NotifEntries{Entries: []model.NotifEntry{entry}}, time.Second, time.Now())
	check(1)
	local = [][]digest.Digest{{current}, {old}}
	check(1)
	local = [][]digest.Digest{{old, current}}
	check(0)
	local = [][]digest.Digest{{current}, nil}
	count, err := testutil.GatherAndCount(reg, "diun_image_local_digest_mismatch")
	require.NoError(t, err)
	require.Zero(t, count)
	local = nil
	count, err = testutil.GatherAndCount(reg, "diun_image_local_digest_mismatch")
	require.NoError(t, err)
	require.Zero(t, count)
	r.SetLocalDigests("docker", func(context.Context) (map[string][][]digest.Digest, error) {
		return nil, context.DeadlineExceeded
	})
	count, err = testutil.GatherAndCount(reg, "diun_image_local_digest_mismatch")
	require.NoError(t, err)
	require.Zero(t, count)
	count, err = testutil.GatherAndCount(reg, "diun_build_info")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
