package app

import (
	"fmt"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/model/provider"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/rs/zerolog/log"
)

type imageJob struct {
	image    provider.Image
	registry *docker.RegistryClient
}

func (di *Diun) procImages() {
	// Iterate images
	for _, img := range di.cfg.Providers.Image {
		regOpts := di.cfg.RegOpts[img.RegOptsID]
		reg, err := docker.NewRegistryClient(docker.RegistryOptions{
			Os:          img.Os,
			Arch:        img.Arch,
			Username:    regOpts.Username,
			Password:    regOpts.Password,
			Timeout:     time.Duration(regOpts.Timeout) * time.Second,
			InsecureTLS: regOpts.InsecureTLS,
		})
		if err != nil {
			log.Error().Err(err).Str("image", img.Name).Msg("Cannot create registry client")
			continue
		}

		image, err := registry.ParseImage(img.Name)
		if err != nil {
			log.Error().Err(err).Str("image", img.Name).Msg("Cannot parse image")
			continue
		}

		di.wg.Add(1)
		err = di.pool.Invoke(imageJob{
			image:    img,
			registry: reg,
		})
		if err != nil {
			log.Error().Err(err).Msgf("Invoking image job")
		}

		if !img.WatchRepo || image.Domain == "" {
			continue
		}

		tags, err := reg.Tags(docker.TagsOptions{
			Image:   image,
			Max:     img.MaxTags,
			Include: img.IncludeTags,
			Exclude: img.ExcludeTags,
		})
		if err != nil {
			log.Error().Err(err).Str("image", image.String()).Msg("Cannot retrieve tags")
			continue
		}

		log.Debug().Str("image", image.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded).",
			tags.Total,
			len(tags.List),
			img.MaxTags,
			tags.NotIncluded,
			tags.Excluded,
		)

		for _, tag := range tags.List {
			img.Name = fmt.Sprintf("%s/%s:%s", image.Domain, image.Path, tag)
			di.wg.Add(1)
			err = di.pool.Invoke(imageJob{
				image:    img,
				registry: reg,
			})
			if err != nil {
				log.Error().Err(err).Msgf("Invoking image job (tag)")
			}
		}
	}
}

func (di *Diun) imageJob(job imageJob) error {
	image, err := registry.ParseImage(job.image.Name)
	if err != nil {
		return err
	}

	if !utl.IsIncluded(image.Tag, job.image.IncludeTags) {
		log.Warn().Str("image", image.String()).Msg("Tag not included")
		return nil
	} else if utl.IsExcluded(image.Tag, job.image.ExcludeTags) {
		log.Warn().Str("image", image.String()).Msg("Tag excluded")
		return nil
	}

	liveManifest, err := job.registry.Manifest(image)
	if err != nil {
		return err
	}

	dbManifest, err := di.db.GetManifest(image)
	if err != nil {
		return err
	}

	status := model.ImageStatusUnchange
	if dbManifest.Name == "" {
		status = model.ImageStatusNew
		log.Info().Str("image", image.String()).Msg("New image found")
	} else if !liveManifest.Created.Equal(*dbManifest.Created) {
		status = model.ImageStatusUpdate
		log.Info().Str("image", image.String()).Msg("Image update found")
	} else {
		log.Debug().Str("image", image.String()).Msg("No changes")
		return nil
	}

	if err := di.db.PutManifest(image, liveManifest); err != nil {
		return err
	}
	log.Debug().Str("image", image.String()).Msg("Manifest saved to database")

	di.notif.Send(model.NotifEntry{
		Status:   status,
		Image:    image,
		Manifest: liveManifest,
	})

	return nil
}
