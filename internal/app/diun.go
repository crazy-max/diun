package app

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/db"
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/crazy-max/diun/pkg/registry"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog/log"
)

// Diun represents an active diun object
type Diun struct {
	cfg    *config.Config
	reg    *registry.Client
	db     *db.Client
	notif  *notif.Client
	locker uint32
}

// New creates new diun instance
func New(cfg *config.Config) (*Diun, error) {
	// Registry client
	regcli, err := registry.New()
	if err != nil {
		return nil, err
	}

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
		reg:   regcli,
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
		image, err := registry.ParseImage(item.Image)
		if err != nil {
			log.Error().Err(err).Str("image", item.Image).Msg("Cannot parse image")
			continue
		}

		opts := &registry.Options{
			Image:       image,
			Username:    item.RegCred.Username,
			Password:    item.RegCred.Password,
			InsecureTLS: item.InsecureTLS,
		}

		if err := di.analyzeImage(item, opts); err != nil {
			log.Error().Err(err).Str("image", opts.Image.String()).Msg("Cannot analyze image")
			continue
		}

		if item.WatchRepo {
			di.analyzeRepo(item, opts)
		}
	}
}

func (di *Diun) analyzeImage(item model.Item, opts *registry.Options) error {
	if !di.isIncluded(opts.Image.Tag, item.IncludeTags) {
		log.Warn().Str("image", opts.Image.String()).Msgf("Tag %s not included", opts.Image.Tag)
		return nil
	} else if di.isExcluded(opts.Image.Tag, item.ExcludeTags) {
		log.Warn().Str("image", opts.Image.String()).Msgf("Tag %s excluded", opts.Image.Tag)
		return nil
	}

	log.Debug().Str("image", opts.Image.String()).Msgf("Analyzing")
	liveAna, err := di.reg.Inspect(opts)
	if err != nil {
		return err
	}

	dbAna, err := di.db.GetAnalysis(opts.Image)
	if err != nil {
		return err
	}

	status := model.ImageStatusUnchange
	if dbAna.Name == "" {
		status = model.ImageStatusNew
		log.Info().Str("image", opts.Image.String()).Msgf("New image found")
	} else if !liveAna.Created.Equal(*dbAna.Created) {
		status = model.ImageStatusUpdate
		log.Info().Str("image", opts.Image.String()).Msgf("Image update found")
	} else {
		log.Debug().Str("image", opts.Image.String()).Msgf("No changes")
		return nil
	}

	if err := di.db.PutAnalysis(opts.Image, liveAna); err != nil {
		return err
	}
	log.Debug().Str("image", opts.Image.String()).Msg("Analysis saved to database")

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Image:    opts.Image,
		Analysis: liveAna,
	})

	return nil
}

func (di *Diun) analyzeRepo(item model.Item, opts *registry.Options) {
	tags, err := di.reg.Tags(opts)
	if err != nil {
		log.Error().Err(err).Str("image", opts.Image.String()).Msg("Cannot retrieve tags")
		return
	}
	log.Debug().Str("image", opts.Image.String()).Msgf("%d tag(s) found", len(tags))

	for _, tag := range tags {
		if tag == opts.Image.Tag {
			continue
		}

		simage := fmt.Sprintf("%s/%s:%s", opts.Image.Domain, opts.Image.Path, tag)
		image, err := registry.ParseImage(simage)
		if err != nil {
			log.Error().Err(err).Str("image", simage).Msg("Cannot parse image")
			continue
		}

		opts := &registry.Options{
			Image:       image,
			Username:    opts.Username,
			Password:    opts.Password,
			InsecureTLS: opts.InsecureTLS,
		}

		if err := di.analyzeImage(item, opts); err != nil {
			log.Error().Err(err).Str("image", image.String()).Msg("Cannot analyze image")
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

func (di *Diun) isIncluded(tag string, includes []string) bool {
	if len(includes) == 0 {
		return true
	}
	for _, include := range includes {
		if utl.MatchString(include, tag) {
			return true
		}
	}
	return false
}

func (di *Diun) isExcluded(tag string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}
	for _, exclude := range excludes {
		if utl.MatchString(exclude, tag) {
			return true
		}
	}
	return false
}

func (di *Diun) trackTime(start time.Time, prefix string) {
	log.Info().Msgf("%s%s", prefix, durafmt.ParseShort(time.Since(start)).String())
}
