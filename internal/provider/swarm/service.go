package swarm

import (
	"reflect"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/docker/docker/api/types/filters"
)

func (c *Client) listServiceImage() []model.Image {
	cli, err := docker.New(docker.Options{
		Endpoint:    c.config.Endpoint,
		APIVersion:  c.config.APIVersion,
		TLSCertPath: c.config.TLSCertsPath,
		TLSVerify:   *c.config.TLSVerify,
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot create Docker client")
		return []model.Image{}
	}

	svcs, err := cli.ServiceList(filters.NewArgs())
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot list Swarm services")
		return []model.Image{}
	}

	var list []model.Image
	for _, svc := range svcs {
		imageName := svc.Spec.TaskTemplate.ContainerSpec.Image
		imageRaw, err := cli.ImageInspectWithRaw(svc.Spec.TaskTemplate.ContainerSpec.Image)
		if err != nil {
			c.logger.Error().Err(err).
				Str("svc_name", svc.Spec.Name).
				Str("ctn_image", imageName).
				Msg("Cannot inspect image")
			continue
		}

		if local := cli.IsLocalImage(imageRaw); local {
			c.logger.Debug().
				Str("svc_name", svc.Spec.Name).
				Str("ctn_image", imageName).
				Msg("Skip locally built image")
			continue
		}

		if dangling := cli.IsDanglingImage(imageRaw); dangling {
			c.logger.Debug().
				Str("svc_name", svc.Spec.Name).
				Str("ctn_image", imageName).
				Msg("Skip dangling image")
			continue
		}

		if cli.IsDigest(imageName) {
			if len(imageRaw.RepoDigests) > 0 {
				c.logger.Debug().
					Str("svc_name", svc.Spec.Name).
					Str("ctn_image", imageName).
					Strs("img_repodigests", imageRaw.RepoDigests).
					Msg("Using first image repo digest available as image name")
				imageName = imageRaw.RepoDigests[0]
			} else {
				c.logger.Debug().
					Str("svc_name", svc.Spec.Name).
					Str("ctn_image", imageName).
					Strs("img_repodigests", imageRaw.RepoDigests).
					Msg("Skip unknown image digest ref")
				continue
			}
		}

		c.logger.Debug().
			Str("svc_name", svc.Spec.Name).
			Interface("svc_labels", svc.Spec.Labels).
			Str("ctn_image", imageName).
			Msg("Validate image")

		image, err := provider.ValidateImage(imageName, svc.Spec.Labels, *c.config.WatchByDefault)
		if err != nil {
			c.logger.Error().Err(err).
				Str("svc_name", svc.Spec.Name).
				Interface("svc_labels", svc.Spec.Labels).
				Str("ctn_image", svc.Spec.TaskTemplate.ContainerSpec.Image).
				Msg("Invalid image")
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			c.logger.Debug().
				Str("svc_name", svc.Spec.Name).
				Interface("svc_labels", svc.Spec.Labels).
				Str("ctn_image", svc.Spec.TaskTemplate.ContainerSpec.Image).
				Msg("Watch disabled")
			continue
		}

		list = append(list, image)
	}

	return list
}
