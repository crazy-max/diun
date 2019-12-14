package app

import (
	"fmt"
	"regexp"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/crazy-max/diun/pkg/utl"
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
	job.RegImage, err = registry.ParseImage(job.Image.Name)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot parse image")
		return
	}

	// Registry options
	regOpts, err := di.getRegOpts(job.Image.RegOptsID)
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
		Os:        "linux",
		Arch:      "amd64",
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

	job.Registry, err = docker.NewRegistryClient(docker.RegistryOptions{
		Os:          job.Image.Os,
		Arch:        job.Image.Arch,
		Username:    regUser,
		Password:    regPassword,
		Timeout:     time.Duration(regOpts.Timeout) * time.Second,
		InsecureTLS: regOpts.InsecureTLS,
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

	if !job.Image.WatchRepo || job.RegImage.Domain == "" {
		return
	}

	tags, err := job.Registry.Tags(docker.TagsOptions{
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
		job.RegImage, err = registry.ParseImage(job.Image.Name)
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

func (di *Diun) runJob(job model.Job) error {
	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.RegImage.String()).
		Logger()

	if !utl.IsIncluded(job.RegImage.Tag, job.Image.IncludeTags) {
		sublog.Warn().Msg("Tag not included")
		return nil
	} else if utl.IsExcluded(job.RegImage.Tag, job.Image.ExcludeTags) {
		sublog.Warn().Msg("Tag excluded")
		return nil
	}

	liveManifest, err := job.Registry.Manifest(job.RegImage)
	if err != nil {
		return err
	}

	dbManifest, err := di.db.GetManifest(job.RegImage)
	if err != nil {
		return err
	}

	status := model.ImageStatusUnchange
	if dbManifest.Name == "" {
		status = model.ImageStatusNew
		sublog.Info().Msg("New image found")
	} else if !liveManifest.Created.Equal(*dbManifest.Created) {
		status = model.ImageStatusUpdate
		sublog.Info().Msg("Image update found")
	} else {
		sublog.Debug().Msg("No changes")
		return nil
	}

	if err := di.db.PutManifest(job.RegImage, liveManifest); err != nil {
		return err
	}
	sublog.Debug().Msg("Manifest saved to database")

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Provider: job.Provider,
		Image:    job.RegImage,
		Manifest: liveManifest,
	})

	return nil
}

func (di *Diun) getRegOpts(id string) (model.RegOpts, error) {
	if id == "" {
		return model.RegOpts{}, nil
	}
	if regopts, ok := di.cfg.RegOpts[id]; ok {
		return regopts, nil
	}
	return model.RegOpts{}, fmt.Errorf("%s not found", id)
}
