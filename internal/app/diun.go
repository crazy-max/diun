package app

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg    *config.Config
	db     *db.Client
	notif  *notif.Client
	locker uint32
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
	defer di.trackTime(time.Now(), "Finished, total time spent: ")

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

		image, err := di.analyzeImage(item.Image, item, reg)
		if err != nil {
			log.Error().Err(err).Str("image", item.Image).Msg("Cannot analyze image")
		}

		if image.Domain != "" && item.WatchRepo {
			di.analyzeRepo(image, item, reg)
		}
	}
}

func (di *Diun) analyzeImage(imageStr string, item model.Item, reg *docker.RegistryClient) (registry.Image, error) {
	image, err := registry.ParseImage(imageStr)
	if err != nil {
		return registry.Image{}, fmt.Errorf("cannot parse image name %s: %v", item.Image, err)
	}

	if !utl.IsIncluded(image.Tag, item.IncludeTags) {
		log.Warn().Str("image", image.String()).Msgf("Tag %s not included", image.Tag)
		return image, nil
	} else if utl.IsExcluded(image.Tag, item.ExcludeTags) {
		log.Warn().Str("image", image.String()).Msgf("Tag %s excluded", image.Tag)
		return image, nil
	}

	log.Debug().Str("image", image.String()).Msgf("Fetching manifest")
	liveManifest, err := reg.Manifest(image)
	if err != nil {
		return image, err
	}
	b, _ := json.MarshalIndent(liveManifest, "", "  ")
	log.Debug().Msg(string(b))

	dbManifest, err := di.db.GetManifest(image)
	if err != nil {
		return image, err
	}

	status := model.ImageStatusUnchange
	if dbManifest.Name == "" {
		status = model.ImageStatusNew
		log.Info().Str("image", image.String()).Msgf("New image found")
	} else if !liveManifest.Created.Equal(*dbManifest.Created) {
		status = model.ImageStatusUpdate
		log.Info().Str("image", image.String()).Msgf("Image update found")
	} else {
		log.Debug().Str("image", image.String()).Msgf("No changes")
		return image, nil
	}

	if err := di.db.PutManifest(image, liveManifest); err != nil {
		return image, err
	}
	log.Debug().Str("image", image.String()).Msg("Manifest saved to database")

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Image:    image,
		Manifest: liveManifest,
	})

	return image, nil
}

func (di *Diun) analyzeRepo(image registry.Image, item model.Item, reg *docker.RegistryClient) {
	tags, err := reg.Tags(docker.TagsOptions{
		Image:   image,
		Max:     item.MaxTags,
		Include: item.IncludeTags,
		Exclude: item.ExcludeTags,
	})
	if err != nil {
		log.Error().Err(err).Str("image", image.String()).Msg("Cannot retrieve tags")
		return
	}
	log.Debug().Str("image", image.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded).",
		tags.Total,
		len(tags.List),
		item.MaxTags,
		tags.NotIncluded,
		tags.Excluded,
	)

	for _, tag := range tags.List {
		imageStr := fmt.Sprintf("%s/%s:%s", image.Domain, image.Path, tag)
		if _, err := di.analyzeImage(imageStr, item, reg); err != nil {
			log.Error().Err(err).Str("image", imageStr).Msg("Cannot analyze image")
			continue
		}
	}
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
