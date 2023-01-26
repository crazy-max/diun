package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/crazy-max/diun/v4/internal/model"
)

var metrics *Metrics

var staleImages = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "diun_stale_image",
		Help: "todo",
	},
	[]string{"image", "container", "label_schema_group"},
)

// Metric is the data points of a single scan
type Metric struct {
	Scanned int
	Updated int
	Failed  int
	Stale   int
}

// Metrics is the handler processing all individual scan metrics
type Metrics struct {
	channel     chan *Metric
	scanned     prometheus.Gauge
	updated     prometheus.Gauge
	failed      prometheus.Gauge
	stale       prometheus.Gauge
	total       prometheus.Counter
	skipped     prometheus.Counter
	staleImages []prometheus.GaugeVec
}

// Register registers metrics for an executed scan
func (metrics *Metrics) Register(metric *Metric) {
	metrics.channel <- metric
}

// Default creates a new metrics handler if none exists, otherwise returns the existing one
func Default() *Metrics {
	if metrics != nil {
		return metrics
	}

	prometheus.Register(staleImages)

	metrics = &Metrics{
		scanned: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "diun_containers_scanned",
			Help: "Number of containers scanned for changes during the last scan",
		}),
		updated: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "diun_containers_updated",
			Help: "Number of containers updated during the last scan",
		}),
		failed: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "diun_containers_failed",
			Help: "Number of containers where update failed during the last scan",
		}),
		stale: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "diun_containers_stale",
			Help: "Number of containers identified that could be updated during the last scan",
		}),
		total: promauto.NewCounter(prometheus.CounterOpts{
			Name: "diun_scans_total",
			Help: "Number of scans since diun started",
		}),
		skipped: promauto.NewCounter(prometheus.CounterOpts{
			Name: "diun_scans_skipped",
			Help: "Number of skipped scans since diun started",
		}),
		channel: make(chan *Metric, 10),
	}

	go metrics.HandleUpdate(metrics.channel)

	return metrics
}

// Register a Notifications Metrics.
func RegisterNotification(s model.NotifEntries) {
	RegisterScan(NewMetric(s))

	for _, item := range s.Entries {
		labels := prometheus.Labels{"image": item.Image.String(), "container": item.ContainerName, "label_schema_group": ""}

		if v, ok := item.ContainerLabels["diun.label-schema.group"]; ok {
			labels = prometheus.Labels{"image": item.Image.String(), "container": item.ContainerName, "label_schema_group": v}
		}

		if item.Status == model.ImageStatusStale {
			staleImages.With(labels).Set(1)
		} else {
			staleImages.With(labels).Set(0)
		}
	}
}

// RegisterScan fetches a metric handler and enqueues a metric
func RegisterScan(metric *Metric) {
	metrics := Default()
	metrics.Register(metric)
}

func NewMetric(s model.NotifEntries) *Metric {
	return &Metric{
		Scanned: s.CountTotal,
		Stale:   s.CountStale,
	}
}

// HandleUpdate dequeue the metric channel and processes it
func (metrics *Metrics) HandleUpdate(channel <-chan *Metric) {
	for change := range channel {
		if change == nil {
			// Update was skipped and rescheduled
			metrics.total.Inc()
			metrics.skipped.Inc()
			metrics.scanned.Set(0)
			metrics.updated.Set(0)
			metrics.failed.Set(0)
			metrics.stale.Set(0)
			continue
		}
		// Update metrics with the new values
		metrics.total.Inc()
		metrics.scanned.Set(float64(change.Scanned))
		metrics.updated.Set(float64(change.Updated))
		metrics.failed.Set(float64(change.Failed))
		metrics.stale.Set(float64(change.Stale))
	}
}
