package docker

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-units"
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
	defer cli.Close()

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
		imageName := ctn.Image
		imageRaw, err := cli.ImageInspectWithRaw(imageName)
		if err != nil {
			c.logger.Error().Err(err).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Msg("Cannot inspect image")
			continue
		}

		if local := cli.IsLocalImage(imageRaw); local {
			c.logger.Debug().
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Msg("Skip locally built image")
			continue
		}

		if dangling := cli.IsDanglingImage(imageRaw); dangling {
			c.logger.Debug().
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Msg("Skip dangling image")
			continue
		}

		if cli.IsDigest(imageName) {
			if len(imageRaw.RepoDigests) > 0 {
				c.logger.Debug().
					Str("ctn_id", ctn.ID).
					Str("ctn_image", imageName).
					Strs("img_repodigests", imageRaw.RepoDigests).
					Msg("Using first image repo digest available as image name")
				imageName = imageRaw.RepoDigests[0]
			} else {
				c.logger.Debug().
					Str("ctn_id", ctn.ID).
					Str("ctn_image", imageName).
					Strs("img_repodigests", imageRaw.RepoDigests).
					Msg("Skip unknown image digest ref")
				continue
			}
		}

		c.logger.Debug().
			Str("ctn_id", ctn.ID).
			Str("ctn_image", imageName).
			Interface("ctn_labels", ctn.Labels).
			Msg("Validate image")
		image, err := provider.ValidateImage(imageName, metadata(ctn), ctn.Labels, *c.config.WatchByDefault, c.defaults)

		if err != nil {
			c.logger.Error().Err(err).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Interface("ctn_labels", ctn.Labels).
				Msg("Invalid image")
			continue
		} else if reflect.DeepEqual(image, model.Image{}) {
			c.logger.Debug().
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Interface("ctn_labels", ctn.Labels).
				Msg("Watch disabled")
			continue
		}

		list = append(list, image)
	}

	return list
}

func metadata(ctn types.Container) map[string]string {
	return map[string]string{
		"ctn_id":        ctn.ID,
		"ctn_names":     formatNames(ctn.Names),
		"ctn_command":   ctn.Command,
		"ctn_createdat": time.Unix(ctn.Created, 0).String(),
		"ctn_state":     ctn.State,
		"ctn_status":    ctn.Status,
		"ctn_size":      formatSize(ctn.SizeRw, ctn.SizeRootFs),
	}
}

func formatNames(names []string) string {
	namesStripPrefix := make([]string, len(names))
	for i, s := range names {
		namesStripPrefix[i] = s[1:]
	}
	return strings.Join(namesStripPrefix, ",")
}

func formatSize(sizeRw, sizeRootFs int64) string {
	srw := units.HumanSizeWithPrecision(float64(sizeRw), 3)
	sv := units.HumanSizeWithPrecision(float64(sizeRootFs), 3)
	sf := srw
	if sizeRootFs > 0 {
		sf = fmt.Sprintf("%s (virtual %s)", srw, sv)
	}
	return sf
}
