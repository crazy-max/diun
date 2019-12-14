package docker

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/rs/zerolog/log"
)

func (c *Client) listContainerImage(elt model.PrdDocker) []model.Image {
	sublog := log.With().
		Str("provider", "docker").
		Str("id", elt.ID).
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

	ctns, err := cli.Containers(ctnFilter)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot list Docker containers")
		return []model.Image{}
	}

	var list []model.Image
	for _, ctn := range ctns {
		image, err := c.containerImage(elt, ctn)
		if err != nil {
			sublog.Error().Err(err).Msgf("Cannot get image for container %s", ctn.ID)
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			sublog.Debug().Msgf("Watch disabled for container %s", ctn.ID)
			continue
		}
		list = append(list, image)
	}

	return list
}

func (c *Client) containerImage(elt model.PrdDocker, ctn types.Container) (img model.Image, err error) {
	img = model.Image{
		Name: ctn.Image,
	}

	if enableStr, ok := ctn.Labels["diun.enable"]; ok {
		enable, err := strconv.ParseBool(enableStr)
		if err != nil {
			return img, fmt.Errorf("cannot parse %s value of label diun.enable", enableStr)
		}
		if !enable {
			return model.Image{}, nil
		}
	} else if !elt.WatchByDefault {
		return model.Image{}, nil
	}

	for key, value := range ctn.Labels {
		switch key {
		case "diun.os":
			img.Os = value
		case "diun.arch":
			img.Arch = value
		case "diun.regopts_id":
			img.RegOptsID = value
		case "diun.watch_repo":
			if img.WatchRepo, err = strconv.ParseBool(value); err != nil {
				return img, fmt.Errorf("cannot parse %s value of label %s", value, key)
			}
		case "diun.max_tags":
			if img.MaxTags, err = strconv.Atoi(value); err != nil {
				return img, fmt.Errorf("cannot parse %s value of label %s", value, key)
			}
		case "diun.include_tags":
			img.IncludeTags = strings.Split(value, ";")
		case "diun.exclude_tags":
			img.ExcludeTags = strings.Split(value, ";")
		}
	}

	return img, nil
}
