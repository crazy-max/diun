package app

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif"
	dockerPrd "github.com/crazy-max/diun/internal/provider/docker"
	staticPrd "github.com/crazy-max/diun/internal/provider/static"
	swarmPrd "github.com/crazy-max/diun/internal/provider/swarm"
	"github.com/hako/durafmt"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg       *config.Config
	cron      *cron.Cron
	db        *db.Client
	notif     *notif.Client
	userAgent string
	jobID     cron.EntryID
	locker    uint32
	pool      *ants.PoolWithFunc
	wg        *sync.WaitGroup
}

// New creates new diun instance
func New(cfg *config.Config, location *time.Location) (*Diun, error) {
	// DB client
	dbcli, err := db.New(cfg.Db)
	if err != nil {
		return nil, err
	}

	// User-Agent
	userAgent := fmt.Sprintf("diun/%s go/%s %s", cfg.App.Version, runtime.Version()[2:], strings.Title(runtime.GOOS))

	// Notification client
	notifcli, err := notif.New(cfg.Notif, cfg.App, userAgent)
	if err != nil {
		return nil, err
	}

	return &Diun{
		cfg: cfg,
		cron: cron.New(cron.WithLocation(location), cron.WithParser(cron.NewParser(
			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor),
		)),
		db:        dbcli,
		notif:     notifcli,
		userAgent: userAgent,
	}, nil
}

// Start starts diun
func (di *Diun) Start() error {
	var err error

	// Migrate db
	err = di.db.Migrate()
	if err != nil {
		return err
	}

	// Run on startup
	di.Run()

	// Init scheduler
	di.jobID, err = di.cron.AddJob(di.cfg.Watch.Schedule, di)
	if err != nil {
		return err
	}
	log.Info().Msgf("Cron initialized with schedule %s", di.cfg.Watch.Schedule)

	// Start scheduler
	di.cron.Start()
	log.Info().Msgf("Next run in %s (%s)",
		durafmt.ParseShort(di.cron.Entry(di.jobID).Next.Sub(time.Now())).String(),
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
			durafmt.ParseShort(di.cron.Entry(di.jobID).Next.Sub(time.Now())).String(),
			di.cron.Entry(di.jobID).Next)
	}

	log.Info().Msg("Cron triggered")
	di.wg = new(sync.WaitGroup)
	di.pool, _ = ants.NewPoolWithFunc(di.cfg.Watch.Workers, func(i interface{}) {
		job := i.(model.Job)
		if err := di.runJob(job); err != nil {
			log.Error().Err(err).
				Str("provider", job.Provider).
				Msg("Cannot run job")
		}
		di.wg.Done()
	})
	defer di.pool.Release()

	// Docker provider
	for _, job := range dockerPrd.New(di.cfg.Providers.Docker).ListJob() {
		di.createJob(job)
	}

	// Swarm provider
	for _, job := range swarmPrd.New(di.cfg.Providers.Swarm).ListJob() {
		di.createJob(job)
	}

	// Static provider
	for _, job := range staticPrd.New(di.cfg.Providers.Static).ListJob() {
		di.createJob(job)
	}

	di.wg.Wait()
}

// Close closes diun
func (di *Diun) Close() {
	if di.cron != nil {
		di.cron.Stop()
	}
	if err := di.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}
