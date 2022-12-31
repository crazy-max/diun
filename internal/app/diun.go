package app

import (
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/cron/v3"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/grpc"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	dockerPrd "github.com/crazy-max/diun/v4/internal/provider/docker"
	dockerfilePrd "github.com/crazy-max/diun/v4/internal/provider/dockerfile"
	filePrd "github.com/crazy-max/diun/v4/internal/provider/file"
	kubernetesPrd "github.com/crazy-max/diun/v4/internal/provider/kubernetes"
	nomadPrd "github.com/crazy-max/diun/v4/internal/provider/nomad"
	swarmPrd "github.com/crazy-max/diun/v4/internal/provider/swarm"
	containerMetrics "github.com/crazy-max/diun/v4/pkg/metrics"
	"github.com/crazy-max/gohealthchecks"
	"github.com/hako/durafmt"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	meta model.Meta
	cfg  *config.Config

	db    *db.Client
	grpc  *grpc.Client
	hc    *gohealthchecks.Client
	notif *notif.Client

	cron   *cron.Cron
	jobID  cron.EntryID
	locker uint32
	pool   *ants.PoolWithFunc
	wg     *sync.WaitGroup
}

// New creates new diun instance
func New(meta model.Meta, cfg *config.Config, grpcAuthority string) (*Diun, error) {
	var err error

	diun := &Diun{
		meta: meta,
		cfg:  cfg,
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		)),
	}

	diun.notif, err = notif.New(cfg.Notif, meta)
	if err != nil {
		return nil, err
	}

	diun.db, err = db.New(*cfg.Db)
	if err != nil {
		return nil, err
	}

	diun.grpc, err = grpc.New(grpcAuthority, diun.db, diun.notif)
	if err != nil {
		return nil, err
	}

	if cfg.Watch.Healthchecks != nil {
		var hcBaseURL *url.URL
		if len(cfg.Watch.Healthchecks.BaseURL) > 0 {
			hcBaseURL, err = url.Parse(cfg.Watch.Healthchecks.BaseURL)
			if err != nil {
				return nil, errors.Wrap(err, "Cannot parse Healthchecks base URL")
			}
		}
		diun.hc = gohealthchecks.NewClient(&gohealthchecks.ClientOptions{
			BaseURL: hcBaseURL,
		})
	}

	return diun, nil
}

// Start starts diun
func (di *Diun) Start() error {
	var err error

	// Migrate db
	err = di.db.Migrate()
	if err != nil {
		return err
	}

	// Start GRPC server
	go func() {
		if err := di.grpc.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start GRPC server")
		}
	}()

	// Run on startup
	di.Run()

	// Init scheduler if defined
	if len(di.cfg.Watch.Schedule) == 0 {
		return nil
	}
	di.jobID, err = di.cron.AddJobWithJitter(di.cfg.Watch.Schedule, di, *di.cfg.Watch.Jitter)
	if err != nil {
		return err
	}
	log.Info().Msgf("Cron initialized with schedule %s", di.cfg.Watch.Schedule)

	// Start scheduler
	di.cron.Start()
	log.Info().Msgf("Next run in %s (%s)",
		durafmt.Parse(time.Until(di.cron.Entry(di.jobID).Next)).LimitFirstN(2).String(),
		di.cron.Entry(di.jobID).Next)

	select {}
}

// Run starts diun
func (di *Diun) Run() {
	if !atomic.CompareAndSwapUint32(&di.locker, 0, 1) {
		log.Warn().Msg("Already running")
		return
	}
	defer atomic.StoreUint32(&di.locker, 0)
	if di.jobID > 0 {
		defer log.Info().Msgf("Next run in %s (%s)",
			durafmt.Parse(time.Until(di.cron.Entry(di.jobID).Next)).LimitFirstN(2).String(),
			di.cron.Entry(di.jobID).Next)
	}

	log.Info().Msg("Cron triggered")
	entries := new(model.NotifEntries)
	di.HealthchecksStart()
	defer di.HealthchecksSuccess(entries)

	di.wg = new(sync.WaitGroup)
	di.pool, _ = ants.NewPoolWithFunc(di.cfg.Watch.Workers, func(i interface{}) {
		job := i.(model.Job)
		entries.Add(di.runJob(job))
		di.wg.Done()
	}, ants.WithLogger(new(logging.AntsLogger)))
	defer di.pool.Release()

	// Docker provider
	for _, job := range dockerPrd.New(di.cfg.Providers.Docker).ListJob() {
		di.createJob(job)
	}

	// Swarm provider
	for _, job := range swarmPrd.New(di.cfg.Providers.Swarm).ListJob() {
		di.createJob(job)
	}

	// Kubernetes provider
	for _, job := range kubernetesPrd.New(di.cfg.Providers.Kubernetes).ListJob() {
		di.createJob(job)
	}

	// File provider
	for _, job := range filePrd.New(di.cfg.Providers.File).ListJob() {
		di.createJob(job)
	}

	// Dockerfile provider
	for _, job := range dockerfilePrd.New(di.cfg.Providers.Dockerfile).ListJob() {
		di.createJob(job)
	}

	// Nomad provider
	for _, job := range nomadPrd.New(di.cfg.Providers.Nomad).ListJob() {
		di.createJob(job)
	}

	di.wg.Wait()
	log.Info().
		Int("added", entries.CountNew).
		Int("updated", entries.CountUpdate).
		Int("unchanged", entries.CountUnchange).
		Int("skipped", entries.CountSkip).
		Int("failed", entries.CountError).
		Int("stale", entries.CountStale).
		Int("total", entries.CountTotal).
		Msg("Jobs completed")

	containerMetrics.RegisterNotification(*entries)

}

// Close closes diun
func (di *Diun) Close() {
	di.HealthchecksFail("Application closed")
	if di.cron != nil {
		di.cron.Stop()
	}
	di.grpc.Stop()
	if err := di.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}
