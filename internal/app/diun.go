package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg       *config.Config
	db        *db.Client
	notif     *notif.Client
	locker    uint32
	collector Collector
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

	log.Info().Msg("Running process")
	var wg sync.WaitGroup
	di.collector = di.StartDispatcher(di.cfg.Watch.Workers)

	// Iterate items
	for _, item := range di.cfg.Items {
		reg, err := docker.NewRegistryClient(docker.RegistryOptions{
			Os:          di.cfg.Watch.Os,
			Arch:        di.cfg.Watch.Arch,
			Username:    item.Registry.Username,
			Password:    item.Registry.Password,
			Timeout:     time.Duration(item.Registry.Timeout) * time.Second,
			InsecureTLS: item.Registry.InsecureTLS,
		})
		if err != nil {
			log.Error().Err(err).Str("image", item.Image).Msg("Cannot create registry client")
			continue
		}

		image, err := registry.ParseImage(item.Image)
		if err != nil {
			log.Error().Err(err).Str("image", item.Image).Msg("Cannot parse image")
			continue
		}

		wg.Add(1)
		di.collector.Job <- Job{
			ImageStr: item.Image,
			Item:     item,
			Reg:      reg,
			Wg:       &wg,
		}

		if image.Domain != "" && item.WatchRepo {
			tags, err := reg.Tags(docker.TagsOptions{
				Image:   image,
				Max:     item.MaxTags,
				Include: item.IncludeTags,
				Exclude: item.ExcludeTags,
			})
			if err != nil {
				log.Error().Err(err).Str("image", image.String()).Msg("Cannot retrieve tags")
				continue
			}

			log.Debug().Str("image", image.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded).",
				tags.Total,
				len(tags.List),
				item.MaxTags,
				tags.NotIncluded,
				tags.Excluded,
			)

			for _, tag := range tags.List {
				wg.Add(1)
				di.collector.Job <- Job{
					ImageStr: fmt.Sprintf("%s/%s:%s", image.Domain, image.Path, tag),
					Item:     item,
					Reg:      reg,
					Wg:       &wg,
				}
			}
		}
	}

	wg.Wait()
}

func (di *Diun) analyze(job Job, workerID int) error {
	defer job.Wg.Done()
	image, err := registry.ParseImage(job.ImageStr)
	if err != nil {
		return err
	}

	if !utl.IsIncluded(image.Tag, job.Item.IncludeTags) {
		log.Warn().Str("image", image.String()).Int("worker_id", workerID).Msg("Tag not included")
		return nil
	} else if utl.IsExcluded(image.Tag, job.Item.ExcludeTags) {
		log.Warn().Str("image", image.String()).Int("worker_id", workerID).Msg("Tag excluded")
		return nil
	}

	liveManifest, err := job.Reg.Manifest(image)
	if err != nil {
		return err
	}
	/*b, _ := json.MarshalIndent(liveManifest, "", "  ")
	log.Debug().Msg(string(b))*/

	dbManifest, err := di.db.GetManifest(image)
	if err != nil {
		return err
	}

	status := model.ImageStatusUnchange
	if dbManifest.Name == "" {
		status = model.ImageStatusNew
		log.Info().Str("image", image.String()).Int("worker_id", workerID).Msg("New image found")
	} else if !liveManifest.Created.Equal(*dbManifest.Created) {
		status = model.ImageStatusUpdate
		log.Info().Str("image", image.String()).Int("worker_id", workerID).Msg("Image update found")
	} else {
		log.Debug().Str("image", image.String()).Int("worker_id", workerID).Msg("No changes")
		return nil
	}

	if err := di.db.PutManifest(image, liveManifest); err != nil {
		return err
	}
	log.Debug().Str("image", image.String()).Int("worker_id", workerID).Msg("Manifest saved to database")

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Image:    image,
		Manifest: liveManifest,
	})

	return nil
}

// Close closes diun
func (di *Diun) Close() {
	if err := di.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}
