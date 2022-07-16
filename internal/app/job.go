package app

import (
	"fmt"
	"regexp"

	"github.com/containers/image/v5/pkg/docker/config"
	"github.com/containers/image/v5/types"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

func (di *Diun) createJob(job model.Job) {
	var err error
	var prvImage registry.Image

	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.Image.Name).
		Logger()

	// Validate image
	prvImage, err = registry.ParseImage(registry.ParseImageOptions{
		Name:   job.Image.Name,
		HubTpl: job.Image.HubTpl,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot parse image")
		return
	}
	job.RegImage = prvImage

	// First check?
	job.FirstCheck, err = di.db.First(job.RegImage)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot check first")
		return
	}

	// Get registry options
	reg, err := di.cfg.RegOpts.Select(job.Image.RegOpt, job.RegImage)
	if err != nil {
		sublog.Warn().Err(err).Msg("Registry options")
	} else if reg != nil {
		sublog.Debug().Str("regopt", reg.Name).Msg("Registry options will be used")
	} else {
		reg = (&model.RegOpt{}).GetDefaults()
	}

	regUser, err := utl.GetSecret(reg.Username, reg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot retrieve username secret for regopts %s", reg.Name)
	}
	regPassword, err := utl.GetSecret(reg.Password, reg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot retrieve password secret for regopts %s", reg.Name)
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

	var auth types.DockerAuthConfig
	if len(regUser) > 0 {
		auth = types.DockerAuthConfig{
			Username: regUser,
			Password: regPassword,
		}
	} else {
		auth, err = config.GetCredentials(nil, job.RegImage.Domain)
		if err != nil {
			sublog.Warn().Err(err).Msg("Error seeking Docker credentials")
		}
	}

	job.Registry, err = registry.New(registry.Options{
		Auth:          auth,
		Timeout:       *reg.Timeout,
		InsecureTLS:   *reg.InsecureTLS,
		UserAgent:     di.meta.UserAgent,
		CompareDigest: *di.cfg.Watch.CompareDigest,
		ImageOs:       job.Image.Platform.OS,
		ImageArch:     job.Image.Platform.Arch,
		ImageVariant:  job.Image.Platform.Variant,
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
		Sort:    job.Image.SortTags,
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
		if prvImage.Tag == tag {
			continue
		}
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

func (di *Diun) runJob(job model.Job) (entry model.NotifEntry) {
	var err error
	entry = model.NotifEntry{
		Status:   model.ImageStatusError,
		Provider: job.Provider,
		Image:    job.RegImage,
	}

	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.RegImage.String()).
		Logger()

	if !utl.IsIncluded(job.RegImage.Tag, job.Image.IncludeTags) {
		entry.Status = model.ImageStatusSkip
		sublog.Debug().Msg("Tag not included")
		return
	} else if utl.IsExcluded(job.RegImage.Tag, job.Image.ExcludeTags) {
		entry.Status = model.ImageStatusSkip
		sublog.Debug().Msg("Tag excluded")
		return
	}

	dbManifest, err := di.db.GetManifest(job.RegImage)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot get manifest from db")
		return
	}

	var updated bool
	entry.Manifest, updated, err = job.Registry.Manifest(job.RegImage, dbManifest)
	if err != nil {
		sublog.Warn().Err(err).Msg("Cannot get remote manifest")
		return
	}

	if len(dbManifest.Name) == 0 {
		entry.Status = model.ImageStatusNew
		sublog.Info().Msg("New image found")
	} else if updated {
		entry.Status = model.ImageStatusUpdate
		sublog.Info().Msg("Image update found")
	} else {
		entry.Status = model.ImageStatusUnchange
		sublog.Debug().Msg("No changes")
	}

	if err := di.db.PutManifest(job.RegImage, entry.Manifest); err != nil {
		sublog.Error().Err(err).Msg("Cannot write manifest to db")
		return
	}
	sublog.Debug().Msg("Manifest saved to database")
	if entry.Status == model.ImageStatusUnchange {
		return
	}

	if job.FirstCheck && !*di.cfg.Watch.FirstCheckNotif {
		sublog.Debug().Msg("Skipping notification (first check)")
		return
	}

	notifyOn := model.NotifyOn(entry.Status)
	if !notifyOn.OneOf(job.Image.NotifyOn) {
		sublog.Debug().Msgf("Skipping notification (%s not part of specified notify status)", entry.Status)
		return
	}

	di.notif.Send(entry)
	return
}
