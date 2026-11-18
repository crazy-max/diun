package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
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

// Recorder records Diun watcher state for Prometheus.
type Recorder struct {
	mu sync.RWMutex

	version      string
	metadataKeys []string

	watchRunsTotal        uint64
	watchSkippedRunsTotal uint64
	lastRunTimestamp      time.Time
	lastRunDuration       time.Duration
	lastRunImages         map[model.ImageStatus]int
	images                map[imageKey]imageState
	localDigests          map[string]LocalDigestSource

	buildInfoDesc                *prometheus.Desc
	watchRunsTotalDesc           *prometheus.Desc
	watchSkippedRunsTotalDesc    *prometheus.Desc
	watchLastRunTimestampDesc    *prometheus.Desc
	watchLastRunDurationDesc     *prometheus.Desc
	watchLastRunImagesDesc       *prometheus.Desc
	imageUpdateAvailableDesc     *prometheus.Desc
	imageLastCheckTimestampDesc  *prometheus.Desc
	imageLastCheckStatusDesc     *prometheus.Desc
	imageCreatedTimestampDesc    *prometheus.Desc
	imageLocalDigestMismatchDesc *prometheus.Desc
}

// LocalDigestSource returns candidate digests per deployed instance; an empty set means unknown.
type LocalDigestSource func(context.Context) (map[string][][]digest.Digest, error)

type imageKey struct {
	provider string
	image    string
	metadata string
}

type imageState struct {
	provider        string
	image           string
	metadataValues  []string
	status          model.ImageStatus
	updateAvailable bool
	lastCheck       time.Time
	created         *time.Time
	digest          digest.Digest
	instances       []digest.Digest
}

// NewRecorder creates a Prometheus registry and recorder for Diun metrics.
func NewRecorder(version string, metadataKeys []string) (*Recorder, *prometheus.Registry) {
	registry := prometheus.NewRegistry()
	recorder := newRecorder(version, metadataKeys)

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		recorder,
	)

	return recorder, registry
}

func newRecorder(version string, metadataKeys []string) *Recorder {
	imageLabels := []string{"provider", "image"}
	imageStatusLabels := []string{"provider", "image", "status"}
	for _, key := range metadataKeys {
		labelName := "metadata_" + key
		imageLabels = append(imageLabels, labelName)
		imageStatusLabels = append(imageStatusLabels, labelName)
	}

	return &Recorder{
		version:                      version,
		metadataKeys:                 append([]string(nil), metadataKeys...),
		lastRunImages:                make(map[model.ImageStatus]int, len(imageStatuses)),
		images:                       make(map[imageKey]imageState),
		imageLocalDigestMismatchDesc: prometheus.NewDesc("diun_image_local_digest_mismatch", "Whether a watched container differs from the last recorded registry digest.", imageLabels, nil),

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
			imageLabels,
			nil,
		),
		imageLastCheckTimestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "last_check_timestamp_seconds"),
			"Unix timestamp of the last completed check for the image.",
			imageLabels,
			nil,
		),
		imageLastCheckStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "last_check_status"),
			"Last check status for the image. The active status has value 1.",
			imageStatusLabels,
			nil,
		),
		imageCreatedTimestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "image", "created_timestamp_seconds"),
			"Unix timestamp of the image manifest creation time reported by the registry.",
			imageLabels,
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
	ch <- r.imageLocalDigestMismatchDesc
}

// Collect sends metric values to Prometheus.
func (r *Recorder) Collect(ch chan<- prometheus.Metric) {
	r.collectRecorded(ch)
	r.mu.RLock()
	sources, images := maps.Clone(r.localDigests), r.images
	r.mu.RUnlock()
	ctx, cancel := context.WithTimeoutCause(context.Background(), 3*time.Second, context.DeadlineExceeded)
	defer cancel()
	var wg sync.WaitGroup
	for name, source := range sources {
		watched := make([]imageState, 0)
		for _, image := range images {
			if image.provider == name && image.digest != "" {
				watched = append(watched, image)
			}
		}
		if len(watched) == 0 {
			continue
		}
		wg.Go(func() {
			local, err := source(ctx)
			if err != nil {
				return
			}
			for _, image := range watched {
				key, err := provider.LocalImageKey(image.image)
				if err != nil {
					continue
				}
				containers := local[key]
				if len(containers) == 0 {
					continue
				}
				value, known := 0.0, true
				for _, digests := range containers {
					if len(digests) == 0 {
						known = false
						continue
					}
					match := slices.Contains(digests, image.digest)
					if name == "kubernetes" {
						for _, dgst := range digests {
							match = match || slices.Contains(image.instances, dgst)
						}
					}
					if !match {
						value = 1
					}
				}
				if known || value == 1 {
					ch <- prometheus.MustNewConstMetric(r.imageLocalDigestMismatchDesc, prometheus.GaugeValue, value, append([]string{image.provider, image.image}, image.metadataValues...)...)
				}
			}
		})
	}
	wg.Wait()
}

// SetLocalDigests configures bounded runtime digest resolution for a provider.
func (r *Recorder) SetLocalDigests(name string, source LocalDigestSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.localDigests == nil {
		r.localDigests = make(map[string]LocalDigestSource)
	}
	r.localDigests[name] = source
}

func (r *Recorder) collectRecorded(ch chan<- prometheus.Metric) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ch <- prometheus.MustNewConstMetric(r.buildInfoDesc, prometheus.GaugeValue, 1, r.version)
	ch <- prometheus.MustNewConstMetric(r.watchRunsTotalDesc, prometheus.CounterValue, float64(r.watchRunsTotal))
	ch <- prometheus.MustNewConstMetric(r.watchSkippedRunsTotalDesc, prometheus.CounterValue, float64(r.watchSkippedRunsTotal))

	if !r.lastRunTimestamp.IsZero() {
		ch <- prometheus.MustNewConstMetric(r.watchLastRunTimestampDesc, prometheus.GaugeValue, float64(r.lastRunTimestamp.Unix()))
		ch <- prometheus.MustNewConstMetric(r.watchLastRunDurationDesc, prometheus.GaugeValue, r.lastRunDuration.Seconds())
		for _, status := range imageStatuses {
			ch <- prometheus.MustNewConstMetric(r.watchLastRunImagesDesc, prometheus.GaugeValue, float64(r.lastRunImages[status]), string(status))
		}
	}

	for _, image := range r.images {
		labelValues := append([]string{image.provider, image.image}, image.metadataValues...)

		updateAvailable := 0.0
		if image.updateAvailable {
			updateAvailable = 1
		}
		ch <- prometheus.MustNewConstMetric(r.imageUpdateAvailableDesc, prometheus.GaugeValue, updateAvailable, labelValues...)
		ch <- prometheus.MustNewConstMetric(r.imageLastCheckTimestampDesc, prometheus.GaugeValue, float64(image.lastCheck.Unix()), labelValues...)
		for _, status := range imageStatuses {
			value := 0.0
			if status == image.status {
				value = 1
			}
			statusLabelValues := append([]string{image.provider, image.image, string(status)}, image.metadataValues...)
			ch <- prometheus.MustNewConstMetric(r.imageLastCheckStatusDesc, prometheus.GaugeValue, value, statusLabelValues...)
		}
		if image.created != nil && !image.created.IsZero() {
			ch <- prometheus.MustNewConstMetric(r.imageCreatedTimestampDesc, prometheus.GaugeValue, float64(image.created.Unix()), labelValues...)
		}
	}
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
			// Snapshot the configured values so the recorder does not retain
			// provider-owned metadata maps.
			metadataValues := make([]string, len(r.metadataKeys))
			for i, mkey := range r.metadataKeys {
				metadataValues[i] = entry.Metadata[mkey]
			}
			key := imageKey{
				provider: entry.Provider,
				image:    imageName,
				// Selected metadata values are part of the Prometheus series identity.
				metadata: fmt.Sprintf("%q", metadataValues),
			}
			// Kubernetes runtimes may report the platform manifest rather than
			// the index digest. Reuse the already-fetched index; no registry I/O.
			var instances []digest.Digest
			if entry.Provider == "kubernetes" {
				var index ocispecs.Index
				if json.Unmarshal(entry.Manifest.Raw, &index) == nil {
					for _, descriptor := range index.Manifests {
						instances = append(instances, descriptor.Digest)
					}
				}
			}
			images[key] = imageState{
				instances:       instances,
				provider:        entry.Provider,
				image:           imageName,
				metadataValues:  metadataValues,
				status:          entry.Status,
				updateAvailable: entry.UpdateAvailable(),
				lastCheck:       completedAt,
				created:         entry.Manifest.Created,
				digest:          entry.Manifest.Digest,
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
