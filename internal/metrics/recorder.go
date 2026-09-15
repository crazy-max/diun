package metrics

import (
	"sync"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/opencontainers/go-digest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const namespace = "diun"

var imageStatuses = []model.ImageStatus{
	model.ImageStatusNew,
	model.ImageStatusUpdate,
	model.ImageStatusUnchange,
	model.ImageStatusSkip,
	model.ImageStatusError,
}

// LocalDigestSource reports the digests its containers are currently
// running, keyed by normalized image reference, using only the local
// container/image store - no registry calls. A reference maps to every
// distinct digest in use for it, since several containers can share one
// image reference and be updated at different times. Providers that can
// answer this register themselves with RegisterLocalDigestSource so
// diun_image_local_outdated can be recomputed at scrape time, independent
// of the watch schedule.
//
// Resolution is keyed by image reference rather than by the container IDs
// seen during the last watch run precisely because applying an update
// replaces the container: the IDs go stale at the one moment the answer
// matters.
type LocalDigestSource interface {
	LocalDigests() map[string][]digest.Digest
}

// Recorder records Diun watcher state for Prometheus.
type Recorder struct {
	mu sync.RWMutex

	version string

	watchRunsTotal        uint64
	watchSkippedRunsTotal uint64
	lastRunTimestamp      time.Time
	lastRunDuration       time.Duration
	lastRunImages         map[model.ImageStatus]int
	images                map[imageKey]imageState

	sourcesMu          sync.RWMutex
	localDigestSources map[string]LocalDigestSource

	buildInfoDesc               *prometheus.Desc
	watchRunsTotalDesc          *prometheus.Desc
	watchSkippedRunsTotalDesc   *prometheus.Desc
	watchLastRunTimestampDesc   *prometheus.Desc
	watchLastRunDurationDesc    *prometheus.Desc
	watchLastRunImagesDesc      *prometheus.Desc
	imageUpdateAvailableDesc    *prometheus.Desc
	imageLastCheckTimestampDesc *prometheus.Desc
	imageLastCheckStatusDesc    *prometheus.Desc
	imageCreatedTimestampDesc   *prometheus.Desc
	imageLocalOutdatedDesc      *prometheus.Desc
}

type imageKey struct {
	provider string
	image    string
}

type imageState struct {
	provider        string
	image           string
	status          model.ImageStatus
	updateAvailable bool
	lastCheck       time.Time
	created         *time.Time
	lastDigest      digest.Digest
}

// NewRecorder creates a Prometheus registry and recorder for Diun metrics.
func NewRecorder(version string) (*Recorder, *prometheus.Registry) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder(version)

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		recorder,
	)

	return recorder, registry
}

func newRecorder(version string) *Recorder {
	return &Recorder{
		version:       version,
		lastRunImages: make(map[model.ImageStatus]int, len(imageStatuses)),
		images:        make(map[imageKey]imageState),

		buildInfoDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "build_info"),
			"Build information for the Diun instance.",
			[]string{"version"},
			nil,
		),
		watchRunsTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "watch", "runs_total"),
			"Total number of completed Diun watch runs.",
			nil,
			nil,
		),
		watchSkippedRunsTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "watch", "skipped_runs_total"),
			"Total number of Diun watch runs skipped because another run was already active.",
			nil,
			nil,
		),
		watchLastRunTimestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "watch", "last_run_timestamp_seconds"),
			"Unix timestamp of the last completed Diun watch run.",
			nil,
			nil,
		),
		watchLastRunDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "watch", "last_run_duration_seconds"),
			"Duration in seconds of the last completed Diun watch run.",
			nil,
			nil,
		),
		watchLastRunImagesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "watch", "last_run_images"),
			"Number of images by status in the last completed Diun watch run.",
			[]string{"status"},
			nil,
		),
		imageUpdateAvailableDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "update_available"),
			"Whether the last check found an actionable update for the image.",
			[]string{"provider", "image"},
			nil,
		),
		imageLastCheckTimestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "last_check_timestamp_seconds"),
			"Unix timestamp of the last completed check for the image.",
			[]string{"provider", "image"},
			nil,
		),
		imageLastCheckStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "last_check_status"),
			"Last check status for the image. The active status has value 1.",
			[]string{"provider", "image", "status"},
			nil,
		),
		imageCreatedTimestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "created_timestamp_seconds"),
			"Unix timestamp of the image manifest creation time reported by the registry.",
			[]string{"provider", "image"},
			nil,
		),
		imageLocalOutdatedDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "local_outdated"),
			"Whether the update Diun last found for the image is still unapplied, i.e. a container is running a digest other than the one Diun recorded. Resolved against the local Docker daemon on every scrape, so it drops to 0 as soon as the image is pulled and the container recreated, without waiting for the next watch run. Omitted when no local container can be resolved for the image.",
			[]string{"provider", "image"},
			nil,
		),
	}
}

// Describe sends metric descriptors to Prometheus.
func (r *Recorder) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.buildInfoDesc
	ch <- r.watchRunsTotalDesc
	ch <- r.watchSkippedRunsTotalDesc
	ch <- r.watchLastRunTimestampDesc
	ch <- r.watchLastRunDurationDesc
	ch <- r.watchLastRunImagesDesc
	ch <- r.imageUpdateAvailableDesc
	ch <- r.imageLastCheckTimestampDesc
	ch <- r.imageLastCheckStatusDesc
	ch <- r.imageCreatedTimestampDesc
	ch <- r.imageLocalOutdatedDesc
}

// Collect sends metric values to Prometheus.
func (r *Recorder) Collect(ch chan<- prometheus.Metric) {
	// Snapshot everything under the lock first, then do the (potentially
	// slow, I/O-bound) local digest resolution and metric emission outside
	// of it - a slow or unreachable Docker daemon during a scrape must not
	// stall a concurrent RecordRun.
	r.mu.RLock()
	version := r.version
	watchRunsTotal := r.watchRunsTotal
	watchSkippedRunsTotal := r.watchSkippedRunsTotal
	lastRunTimestamp := r.lastRunTimestamp
	lastRunDuration := r.lastRunDuration
	lastRunImages := r.lastRunImages
	images := make(map[imageKey]imageState, len(r.images))
	for k, v := range r.images {
		images[k] = v
	}
	r.mu.RUnlock()

	ch <- prometheus.MustNewConstMetric(r.buildInfoDesc, prometheus.GaugeValue, 1, version)
	ch <- prometheus.MustNewConstMetric(r.watchRunsTotalDesc, prometheus.CounterValue, float64(watchRunsTotal))
	ch <- prometheus.MustNewConstMetric(r.watchSkippedRunsTotalDesc, prometheus.CounterValue, float64(watchSkippedRunsTotal))

	if !lastRunTimestamp.IsZero() {
		ch <- prometheus.MustNewConstMetric(r.watchLastRunTimestampDesc, prometheus.GaugeValue, float64(lastRunTimestamp.Unix()))
		ch <- prometheus.MustNewConstMetric(r.watchLastRunDurationDesc, prometheus.GaugeValue, lastRunDuration.Seconds())
		for _, status := range imageStatuses {
			ch <- prometheus.MustNewConstMetric(r.watchLastRunImagesDesc, prometheus.GaugeValue, float64(lastRunImages[status]), string(status))
		}
	}

	localDigests := r.resolveLocalDigests(images)

	for key, image := range images {
		updateAvailable := 0.0
		if image.updateAvailable {
			updateAvailable = 1
		}
		ch <- prometheus.MustNewConstMetric(r.imageUpdateAvailableDesc, prometheus.GaugeValue, updateAvailable, image.provider, image.image)
		ch <- prometheus.MustNewConstMetric(r.imageLastCheckTimestampDesc, prometheus.GaugeValue, float64(image.lastCheck.Unix()), image.provider, image.image)
		for _, status := range imageStatuses {
			value := 0.0
			if status == image.status {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(r.imageLastCheckStatusDesc, prometheus.GaugeValue, value, image.provider, image.image, string(status))
		}
		if image.created != nil && !image.created.IsZero() {
			ch <- prometheus.MustNewConstMetric(r.imageCreatedTimestampDesc, prometheus.GaugeValue, float64(image.created.Unix()), image.provider, image.image)
		}
		// Only meaningful when there's a recorded registry digest to
		// compare against; without one we haven't checked, which is not
		// the same as "up to date", so the series is omitted.
		if dgsts, ok := localDigests[key]; ok && len(dgsts) > 0 && len(image.lastDigest) > 0 {
			outdated := 0.0
			// Gated on updateAvailable so that pulling before Diun's next
			// registry poll reads as up to date rather than outdated: if
			// the last poll found nothing actionable, a local digest that
			// differs from the recorded one is ahead of Diun, not behind
			// it. When the poll did flag an update, any container still
			// off the recorded digest hasn't had it applied yet.
			if image.updateAvailable {
				for _, dgst := range dgsts {
					if dgst != image.lastDigest {
						outdated = 1
						break
					}
				}
			}
			ch <- prometheus.MustNewConstMetric(r.imageLocalOutdatedDesc, prometheus.GaugeValue, outdated, image.provider, image.image)
		}
	}
}

// resolveLocalDigests asks each registered LocalDigestSource what its
// containers are running right now, and keeps the answers that match an
// image Diun currently knows about. It performs no registry calls itself;
// each source talks to its own local daemon only. Images whose provider
// has no registered source, or that no local container is running, are
// left out of the result.
func (r *Recorder) resolveLocalDigests(images map[imageKey]imageState) map[imageKey][]digest.Digest {
	providers := make(map[string]struct{})
	for _, img := range images {
		providers[img.provider] = struct{}{}
	}
	if len(providers) == 0 {
		return nil
	}

	r.sourcesMu.RLock()
	sources := make(map[string]LocalDigestSource, len(r.localDigestSources))
	for k, v := range r.localDigestSources {
		sources[k] = v
	}
	r.sourcesMu.RUnlock()

	result := make(map[imageKey][]digest.Digest)
	for provider := range providers {
		source, ok := sources[provider]
		if !ok {
			continue
		}
		for image, digests := range source.LocalDigests() {
			key := imageKey{provider: provider, image: image}
			if _, watched := images[key]; !watched {
				continue
			}
			result[key] = digests
		}
	}
	return result
}

// RegisterLocalDigestSource wires a provider-specific LocalDigestSource
// into the recorder, keyed by provider name (e.g. "docker"). Safe to call
// before or after the metrics server starts serving.
func (r *Recorder) RegisterLocalDigestSource(provider string, source LocalDigestSource) {
	if r == nil || source == nil {
		return
	}

	r.sourcesMu.Lock()
	defer r.sourcesMu.Unlock()

	if r.localDigestSources == nil {
		r.localDigestSources = make(map[string]LocalDigestSource)
	}
	r.localDigestSources[provider] = source
}

// RecordRun records a completed Diun watch run.
func (r *Recorder) RecordRun(entries *model.NotifEntries, duration time.Duration, completedAt time.Time) {
	if r == nil {
		return
	}
	if completedAt.IsZero() {
		completedAt = time.Now()
	}

	lastRunImages := make(map[model.ImageStatus]int, len(imageStatuses))
	images := make(map[imageKey]imageState)
	if entries != nil {
		for _, entry := range entries.Entries {
			lastRunImages[entry.Status]++

			imageName := entry.Image.String()
			if imageName == "" {
				continue
			}
			key := imageKey{
				provider: entry.Provider,
				image:    imageName,
			}
			images[key] = imageState{
				provider:        entry.Provider,
				image:           imageName,
				status:          entry.Status,
				updateAvailable: entry.UpdateAvailable(),
				lastCheck:       completedAt,
				created:         entry.Manifest.Created,
				lastDigest:      entry.Manifest.Digest,
			}
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.watchRunsTotal++
	r.lastRunTimestamp = completedAt
	r.lastRunDuration = duration
	r.lastRunImages = lastRunImages
	r.images = images
}

// RecordSkippedRun records a watch run skipped because another run is active.
func (r *Recorder) RecordSkippedRun() {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.watchSkippedRunsTotal++
}
