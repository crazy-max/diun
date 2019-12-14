package docker

import (
	"fmt"
	"reflect"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/docker/docker/api/types/filters"
	"github.com/rs/zerolog/log"
)

func (c *Client) listContainerImage(id string, elt model.PrdDocker) []model.Image {
	sublog := log.With().
		Str("provider", fmt.Sprintf("docker-%s", id)).
		Logger()

	cli, err := docker.NewClient(elt.Endpoint, elt.ApiVersion, elt.TLSCertsPath, elt.TLSVerify)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot create Docker client")
		return []model.Image{}
	}

	ctnFilter := filters.NewArgs()
	ctnFilter.Add("status", "running")
	if elt.WatchStopped {
		ctnFilter.Add("status", "created")
		ctnFilter.Add("status", "exited")
	}

	ctns, err := cli.ContainerList(ctnFilter)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot list Docker containers")
		return []model.Image{}
	}

	var list []model.Image
	for _, ctn := range ctns {
		image, err := provider.ValidateContainerImage(ctn.Image, ctn.Labels, elt.WatchByDefault)
		if err != nil {
			sublog.Error().Err(err).Msgf("Cannot get image from container %s", ctn.ID)
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			sublog.Debug().Msgf("Watch disabled for container %s", ctn.ID)
			continue
		}
		list = append(list, image)
	}

	return list
}
