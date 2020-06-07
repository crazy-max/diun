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
		local, _ := cli.IsLocalImage(svc.Spec.TaskTemplate.ContainerSpec.Image)
		if local {
			c.logger.Debug().Msgf("Skip locally built image for service %s", svc.Spec.Name)
			continue
		}
		image, err := provider.ValidateContainerImage(svc.Spec.TaskTemplate.ContainerSpec.Image, svc.Spec.Labels, *c.config.WatchByDefault)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Cannot get image from service %s", svc.Spec.Name)
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			c.logger.Debug().Msgf("Watch disabled for service %s", svc.Spec.Name)
			continue
		}
		list = append(list, image)
	}

	return list
}
