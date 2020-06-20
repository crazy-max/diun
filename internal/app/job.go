package app

import (
	"fmt"
	"regexp"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

func (di *Diun) createJob(job model.Job) {
	var err error

	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.Image.Name).
		Logger()

	// Validate image
	job.RegImage, err = registry.ParseImage(registry.ParseImageOptions{
		Name:   job.Image.Name,
		HubTpl: job.Image.HubTpl,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot parse image")
		return
	}

	// First check?
	job.FirstCheck, err = di.db.First(job.RegImage)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot check first")
		return
	}

	// Registry options
	regOpts, err := di.cfg.GetRegOpts(job.Image.RegOptsID)
	if err != nil {
		sublog.Warn().Err(err).Msg("Registry options")
	}

	regUser, err := utl.GetSecret(regOpts.Username, regOpts.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot retrieve username secret for regopts %s", job.Image.RegOptsID)
	}
	regPassword, err := utl.GetSecret(regOpts.Password, regOpts.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot retrieve password secret for regopts %s", job.Image.RegOptsID)
	}

	// Set defaults
	if err := mergo.Merge(&job.Image, model.Image{
		Platform:  model.ImagePlatform{},
		WatchRepo: false,
		MaxTags:   0,
	}); err != nil {
		sublog.Error().Err(err).Msg("Cannot set default values")
		return
	}

	// Validate include/exclude tags
	for _, includeTag := range job.Image.IncludeTags {
		if _, err := regexp.Compile(includeTag); err != nil {
			sublog.Error().Err(err).Msg("Include tag regex '%s' cannot compile")
			return
		}
	}
	for _, excludeTag := range job.Image.ExcludeTags {
		if _, err := regexp.Compile(excludeTag); err != nil {
			sublog.Error().Err(err).Msg("Exclude tag regex '%s' cannot compile")
			return
		}
	}

	job.Registry, err = registry.New(registry.Options{
		Username:     regUser,
		Password:     regPassword,
		Timeout:      *regOpts.Timeout,
		InsecureTLS:  *regOpts.InsecureTLS,
		UserAgent:    di.meta.UserAgent,
		ImageOs:      job.Image.Platform.Os,
		ImageArch:    job.Image.Platform.Arch,
		ImageVariant: job.Image.Platform.Variant,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot create registry client")
		return
	}

	di.wg.Add(1)
	err = di.pool.Invoke(job)
	if err != nil {
		sublog.Error().Err(err).Msgf("Invoking job")
	}

	if !job.Image.WatchRepo || len(job.RegImage.Domain) == 0 {
		return
	}

	tags, err := job.Registry.Tags(registry.TagsOptions{
		Image:   job.RegImage,
		Max:     job.Image.MaxTags,
		Include: job.Image.IncludeTags,
		Exclude: job.Image.ExcludeTags,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot list tags from registry")
		return
	}

	log.Debug().Str("image", job.RegImage.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded).",
		tags.Total,
		len(tags.List),
		job.Image.MaxTags,
		tags.NotIncluded,
		tags.Excluded,
	)

	for _, tag := range tags.List {
		job.Image.Name = fmt.Sprintf("%s/%s:%s", job.RegImage.Domain, job.RegImage.Path, tag)
		job.RegImage, err = registry.ParseImage(registry.ParseImageOptions{
			Name:   job.Image.Name,
			HubTpl: job.Image.HubTpl,
		})
		if err != nil {
			sublog.Error().Err(err).Msg("Cannot parse image (tag)")
			continue
		}
		di.wg.Add(1)
		err = di.pool.Invoke(job)
		if err != nil {
			sublog.Error().Err(err).Msgf("Invoking job (tag)")
		}
	}
}

func (di *Diun) runJob(job model.Job) {
	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.RegImage.String()).
		Logger()

	if !utl.IsIncluded(job.RegImage.Tag, job.Image.IncludeTags) {
		sublog.Warn().Msg("Tag not included")
		return
	} else if utl.IsExcluded(job.RegImage.Tag, job.Image.ExcludeTags) {
		sublog.Warn().Msg("Tag excluded")
		return
	}

	liveManifest, err := job.Registry.Manifest(job.RegImage)
	if err != nil {
		sublog.Warn().Err(err).Msg("Cannot get remote manifest")
		return
	}

	dbManifest, err := di.db.GetManifest(job.RegImage)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot get manifest from db")
		return
	}

	status := model.ImageStatusUnchange
	if len(dbManifest.Name) == 0 {
		status = model.ImageStatusNew
		sublog.Info().Msg("New image found")
	} else if !liveManifest.Created.Equal(*dbManifest.Created) {
		status = model.ImageStatusUpdate
		sublog.Info().Msg("Image update found")
	} else {
		sublog.Debug().Msg("No changes")
		return
	}

	if err := di.db.PutManifest(job.RegImage, liveManifest); err != nil {
		sublog.Error().Err(err).Msg("Cannot write manifest to db")
		return
	}
	sublog.Debug().Msg("Manifest saved to database")

	if job.FirstCheck && !*di.cfg.Watch.FirstCheckNotif {
		sublog.Debug().Msg("Skipping notification (first check)")
		return
	}

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Provider: job.Provider,
		Image:    job.RegImage,
		Manifest: liveManifest,
	})
}
