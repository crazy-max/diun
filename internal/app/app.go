package app

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/cron/v3"
	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/grpc"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/crazy-max/diun/v4/internal/metrics"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	"github.com/crazy-max/diun/v4/internal/provider"
	dockerPrd "github.com/crazy-max/diun/v4/internal/provider/docker"
	dockerfilePrd "github.com/crazy-max/diun/v4/internal/provider/dockerfile"
	filePrd "github.com/crazy-max/diun/v4/internal/provider/file"
	kubernetesPrd "github.com/crazy-max/diun/v4/internal/provider/kubernetes"
	nomadPrd "github.com/crazy-max/diun/v4/internal/provider/nomad"
	swarmPrd "github.com/crazy-max/diun/v4/internal/provider/swarm"
	"github.com/dromara/carbon/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Diun struct {
	meta model.Meta
	cfg  *config.Config

	db            *db.Client
	grpc          *grpc.Client
	hc            *healthchecksClient
	metrics       *metrics.Recorder
	metricsServer *metrics.Server
	notif         *notif.Client

	cron   *cron.Cron
	jobID  cron.EntryID
	locker uint32
	pool   *ants.PoolWithFunc
	wg     *sync.WaitGroup
}

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
		diun.hc, err = newHealthchecksClient(cfg.Watch.Healthchecks)
		if err != nil {
			return nil, err
		}
	}
	if cfg.Metrics != nil && cfg.Metrics.Enabled != nil && *cfg.Metrics.Enabled {
		recorder, registry := metrics.NewRecorder(meta.Version)
		diun.metrics = recorder
		diun.metricsServer, err = metrics.NewServer(cfg.Metrics, registry)
		if err != nil {
			return nil, err
		}
	}

	return diun, nil
}

func (di *Diun) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	if err := di.db.Migrate(); err != nil {
		return err
	}

	lis, err := di.grpc.Listen()
	if err != nil {
		return err
	}

	var metricsLis net.Listener
	if di.metricsServer != nil {
		metricsLis, err = di.metricsServer.Listen()
		if err != nil {
			if closeErr := lis.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("Cannot close gRPC listener")
			}
			return err
		}
	}

	defer func() {
		if di.cron != nil {
			<-di.cron.Stop().Done()
		}
		if di.metricsServer != nil {
			shutdownCtx, cancel := context.WithTimeoutCause(context.Background(), 5*time.Second, errors.New("Prometheus metrics server shutdown timed out"))
			defer cancel()
			if err := di.metricsServer.Shutdown(shutdownCtx); err != nil {
				log.Warn().Err(err).Msg("Cannot stop Prometheus metrics server")
			}
		}
		di.grpc.Stop()
		if err := di.db.Close(); err != nil {
			log.Warn().Err(err).Msg("Cannot close database")
		}
	}()

	serverErrCh := make(chan error, 2)
	go func() {
		serverErrCh <- errors.Wrap(di.grpc.Serve(lis), "gRPC server failed")
	}()
	if di.metricsServer != nil {
		go func() {
			serverErrCh <- errors.Wrap(di.metricsServer.Serve(metricsLis), "Prometheus metrics server failed")
		}()
	}

	if *di.cfg.Watch.RunOnStartup {
		di.Run()
	}

	if len(di.cfg.Watch.Schedule) == 0 {
		return nil
	}
	di.jobID, err = di.cron.AddJobWithJitter(di.cfg.Watch.Schedule, di, *di.cfg.Watch.Jitter)
	if err != nil {
		return err
	}
	log.Info().Msgf("Cron initialized with schedule %s", di.cfg.Watch.Schedule)

	di.cron.Start()
	log.Info().Msgf("Next run in %s (%s)",
		carbon.CreateFromStdTime(di.cron.Entry(di.jobID).Next).DiffAbsInString(),
		di.cron.Entry(di.jobID).Next)

	select {
	case <-ctx.Done():
		di.HealthchecksFail("Application closed")
		return nil
	case err := <-serverErrCh:
		di.HealthchecksFail("Application closed")
		return err
	}
}

func (di *Diun) Run() {
	if !atomic.CompareAndSwapUint32(&di.locker, 0, 1) {
		if di.metrics != nil {
			di.metrics.RecordSkippedRun()
		}
		log.Warn().Msg("Already running")
		return
	}
	startedAt := time.Now()
	defer atomic.StoreUint32(&di.locker, 0)
	if di.jobID > 0 {
		defer log.Info().Msgf("Next run in %s (%s)",
			carbon.CreateFromStdTime(di.cron.Entry(di.jobID).Next).DiffAbsInString(),
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

	provider.WalkJobs(di.createJob,
		dockerPrd.New(di.cfg.Providers.Docker, di.cfg.Defaults),
		swarmPrd.New(di.cfg.Providers.Swarm, di.cfg.Defaults),
		kubernetesPrd.New(di.cfg.Providers.Kubernetes, di.cfg.Defaults),
		filePrd.New(di.cfg.Providers.File, di.cfg.Defaults),
		dockerfilePrd.New(di.cfg.Providers.Dockerfile, di.cfg.Defaults),
		nomadPrd.New(di.cfg.Providers.Nomad, di.cfg.Defaults),
	)

	di.wg.Wait()
	completedAt := time.Now()
	if di.metrics != nil {
		di.metrics.RecordRun(entries, completedAt.Sub(startedAt), completedAt)
	}
	log.Info().
		Int("added", entries.CountNew).
		Int("updated", entries.CountUpdate).
		Int("unchanged", entries.CountUnchange).
		Int("skipped", entries.CountSkip).
		Int("failed", entries.CountError).
		Msg("Jobs completed")
}
