package app

import (
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/db"
	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif"
	dockerPrd "github.com/crazy-max/diun/v4/internal/provider/docker"
	filePrd "github.com/crazy-max/diun/v4/internal/provider/file"
	kubernetesPrd "github.com/crazy-max/diun/v4/internal/provider/kubernetes"
	swarmPrd "github.com/crazy-max/diun/v4/internal/provider/swarm"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/gohealthchecks"
	"github.com/hako/durafmt"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	meta   model.Meta
	cfg    *config.Config
	cron   *cron.Cron
	db     *db.Client
	hc     *gohealthchecks.Client
	notif  *notif.Client
	jobID  cron.EntryID
	locker uint32
	pool   *ants.PoolWithFunc
	wg     *sync.WaitGroup
}

// New creates new diun instance
func New(meta model.Meta, cli model.Cli, cfg *config.Config) (*Diun, error) {
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

	if !cli.TestNotif {
		diun.db, err = db.New(*cfg.Db)
		if err != nil {
			return nil, err
		}
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

	di.wg.Wait()
	log.Info().
		Int("added", entries.CountNew).
		Int("updated", entries.CountUpdate).
		Int("unchanged", entries.CountUnchange).
		Int("failed", entries.CountError).
		Msg("Jobs completed")
}

// Close closes diun
func (di *Diun) Close() {
	di.HealthchecksFail("Application closed")
	if di.cron != nil {
		di.cron.Stop()
	}
	if err := di.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}

// TestNotif test the notification settings
func (di *Diun) TestNotif() {
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", "2020-03-26T12:23:56Z")
	image, _ := registry.ParseImage(registry.ParseImageOptions{
		Name: "diun/testnotif:latest",
	})
	image.HubLink = ""

	log.Info().Msg("Testing notification settings...")
	di.notif.Send(model.NotifEntry{
		Status:   "new",
		Provider: "file",
		Image:    image,
		Manifest: registry.Manifest{
			Name:          "diun/testnotif",
			Tag:           "latest",
			MIMEType:      "application/vnd.docker.distribution.manifest.list.v2+json",
			Digest:        "sha256:216e3ae7de4ca8b553eb11ef7abda00651e79e537e85c46108284e5e91673e01",
			Created:       &createdAt,
			DockerVersion: "",
			Labels: map[string]string{
				"maintainer":                      "CrazyMax",
				"org.label-schema.build-date":     "2020-03-26T12:23:56Z",
				"org.label-schema.description":    "Docker image update notifier",
				"org.label-schema.name":           "Diun",
				"org.label-schema.schema-version": "1.0",
				"org.label-schema.url":            "https://github.com/crazy-max/diun",
				"org.label-schema.vcs-ref":        "e13f097c",
				"org.label-schema.vcs-url":        "https://github.com/crazy-max/diun",
				"org.label-schema.vendor":         "CrazyMax",
				"org.label-schema.version":        "x.x.x",
			},
			Layers: []string{
				"sha256:aad63a9339440e7c3e1fff2b988991b9bfb81280042fa7f39a5e327023056819",
				"sha256:166c6f165b73185ede72415d780538a55c0c8e854bd177925bc007193e5b0d1b",
				"sha256:e05682efa9cc9d6239b2b9252fe0dc1e58d6e1585679733bb94a6549d49e9b10",
				"sha256:c6a5bfed445b3ed7e85523cd73c6532ac9f9b72bb588ca728fd5b33987ca6538",
				"sha256:df2140efb8abeb727ef0b27ff158b7010a7941eb1cfdade505f510a6e1eaf016",
			},
			Platform: "linux/amd64",
		},
	})
}
