package app

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/notif"
	"github.com/hako/durafmt"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg    *config.Config
	cron   *cron.Cron
	db     *db.Client
	notif  *notif.Client
	jobID  cron.EntryID
	locker uint32
	pool   *ants.PoolWithFunc
	wg     *sync.WaitGroup
}

// New creates new diun instance
func New(cfg *config.Config, location *time.Location) (*Diun, error) {
	// DB client
	dbcli, err := db.New(cfg.Db)
	if err != nil {
		return nil, err
	}

	// Notification client
	notifcli, err := notif.New(cfg.Notif, cfg.App)
	if err != nil {
		return nil, err
	}

	return &Diun{
		cfg:   cfg,
		cron:  cron.New(cron.WithLocation(location), cron.WithSeconds()),
		db:    dbcli,
		notif: notifcli,
	}, nil
}

// Start starts diun
func (di *Diun) Start() error {
	var err error

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

// Run runs diun process
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

	log.Info().Msg("Starting Diun...")
	di.wg = new(sync.WaitGroup)
	di.pool, _ = ants.NewPoolWithFunc(di.cfg.Watch.Workers, func(i interface{}) {
		var err error
		switch t := i.(type) {
		case imageJob:
			err = di.imageJob(t)
			if err != nil {
				log.Error().Err(err).Msg("Job image error")
			}
		}
		di.wg.Done()
	})
	defer di.pool.Release()

	di.procImages()
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
