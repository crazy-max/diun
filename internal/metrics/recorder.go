package metrics

import (
	"sync"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
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

	version string

	watchRunsTotal        uint64
	watchSkippedRunsTotal uint64
	lastRunTimestamp      time.Time
	lastRunDuration       time.Duration
	lastRunImages         map[model.ImageStatus]int
	images                map[imageKey]imageState

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
}

// Collect sends metric values to Prometheus.
func (r *Recorder) Collect(ch chan<- prometheus.Metric) {
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
