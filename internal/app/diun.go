package app

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/notif"
	"github.com/hako/durafmt"
	"github.com/panjf2000/ants"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg    *config.Config
	db     *db.Client
	notif  *notif.Client
	locker uint32
	pool   *ants.PoolWithFunc
	wg     *sync.WaitGroup
}

// New creates new diun instance
func New(cfg *config.Config) (*Diun, error) {
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
		db:    dbcli,
		notif: notifcli,
	}, nil
}

// Run starts diun process
func (di *Diun) Run() {
	if !atomic.CompareAndSwapUint32(&di.locker, 0, 1) {
		log.Warn().Msg("Already running")
		return
	}
	defer atomic.StoreUint32(&di.locker, 0)

	start := time.Now()
	defer di.trackTime(start, "Finished, total time spent: ")

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
			err = di.imageRepoJob(t)
			if err != nil {
				log.Error().Err(err).Msg("Job image repo error")
			}
		}
		di.wg.Done()
	})
	defer func() {
		if err := di.pool.Release(); err != nil {
			log.Warn().Err(err).Msg("Cannot release pool")
		}
	}()

	di.procImages()
	di.wg.Wait()
}

// Close closes diun
func (di *Diun) Close() {
	if err := di.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}

func (di *Diun) trackTime(start time.Time, prefix string) {
	log.Info().Msgf("%s%s", prefix, durafmt.ParseShort(time.Since(start)).String())
}
