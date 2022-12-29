package swarm

import (
	"reflect"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
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
	defer cli.Close()

	svcs, err := cli.ServiceList(filters.NewArgs())
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot list Swarm services")
		return []model.Image{}
	}

	var list []model.Image
	for _, svc := range svcs {
		c.logger.Debug().
			Str("svc_name", svc.Spec.Name).
			Interface("svc_labels", svc.Spec.Labels).
			Str("ctn_image", svc.Spec.TaskTemplate.ContainerSpec.Image).
			Msg("Validate image")

		image, err := provider.ValidateImage(svc.Spec.TaskTemplate.ContainerSpec.Image, metadata(svc), svc.Spec.Labels, *c.config.WatchByDefault)
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

func metadata(svc swarm.Service) map[string]string {
	return map[string]string{
		"svc_id":        svc.ID,
		"svc_createdat": svc.CreatedAt.String(),
		"svc_updatedat": svc.UpdatedAt.String(),
		"ctn_name":      svc.Spec.Name,
	}
}
