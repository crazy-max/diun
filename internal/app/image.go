package app

import (
	"fmt"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/rs/zerolog/log"
)

type imageJob struct {
	origin   bool
	image    model.Image
	registry *docker.RegistryClient
}

func (di *Diun) procImages() {
	// Iterate images
	for _, img := range di.cfg.Image {
		reg, err := docker.NewRegistryClient(docker.RegistryOptions{
			Os:          di.cfg.Watch.Os,
			Arch:        di.cfg.Watch.Arch,
			Username:    img.RegOpts.Username,
			Password:    img.RegOpts.Password,
			Timeout:     time.Duration(img.RegOpts.Timeout) * time.Second,
			InsecureTLS: img.RegOpts.InsecureTLS,
		})
		if err != nil {
			log.Error().Err(err).Str("image", img.Name).Msg("Cannot create registry client")
			continue
		}

		di.wg.Add(1)
		err = di.pool.Invoke(imageJob{
			origin:   true,
			image:    img,
			registry: reg,
		})
		if err != nil {
			log.Error().Err(err).Msgf("Invoking image job")
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
	/*b, _ := json.MarshalIndent(liveManifest, "", "  ")
	log.Debug().Msg(string(b))*/

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

func (di *Diun) imageRepoJob(job imageJob) error {
	image, err := registry.ParseImage(job.image.Name)
	if err != nil {
		return err
	}

	if !job.origin || image.Domain == "" || !job.image.WatchRepo {
		return nil
	}

	tags, err := job.registry.Tags(docker.TagsOptions{
		Image:   image,
		Max:     job.image.MaxTags,
		Include: job.image.IncludeTags,
		Exclude: job.image.ExcludeTags,
	})
	if err != nil {
		return err
	}

	log.Debug().Str("image", image.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded).",
		tags.Total,
		len(tags.List),
		job.image.MaxTags,
		tags.NotIncluded,
		tags.Excluded,
	)

	job.origin = false
	for _, tag := range tags.List {
		job.image.Name = fmt.Sprintf("%s/%s:%s", image.Domain, image.Path, tag)
		di.wg.Add(1)
		err = di.pool.Invoke(job)
		if err != nil {
			log.Error().Err(err).Msgf("Invoking repo image job")
		}
	}

	return nil
}
