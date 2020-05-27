package docker

import (
	"reflect"

	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/internal/provider"
	"github.com/crazy-max/diun/v3/pkg/docker"
	"github.com/docker/docker/api/types/filters"
)

func (c *Client) listContainerImage() []model.Image {
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

	ctnFilter := filters.NewArgs()
	ctnFilter.Add("status", "running")
	if *c.config.WatchStopped {
		ctnFilter.Add("status", "created")
		ctnFilter.Add("status", "exited")
	}

	ctns, err := cli.ContainerList(ctnFilter)
	if err != nil {
		c.logger.Error().Err(err).Msg("Cannot list Docker containers")
		return []model.Image{}
	}

	var list []model.Image
	for _, ctn := range ctns {
		local, err := cli.IsLocalImage(ctn.Image)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Cannot inspect image from container %s", ctn.ID)
			continue
		} else if local {
			c.logger.Debug().Msgf("Skip locally built image for container %s", ctn.ID)
			continue
		}
		image, err := provider.ValidateContainerImage(ctn.Image, ctn.Labels, *c.config.WatchByDefault)
		if err != nil {
			c.logger.Error().Err(err).Msgf("Cannot get image from container %s", ctn.ID)
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			c.logger.Debug().Msgf("Watch disabled for container %s", ctn.ID)
			continue
		}
		list = append(list, image)
	}

	return list
}
