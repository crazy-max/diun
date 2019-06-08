package docker

import (
	"github.com/containers/image/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
)

type Tags []string

// Tags returns tags of a Docker repository
func (c *RegistryClient) Tags(image registry.Image) (Tags, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgCls, err := c.newImage(ctx, image.String())
	if err != nil {
		return nil, err
	}
	defer imgCls.Close()

	tags, err := docker.GetRepositoryTags(ctx, c.sysCtx, imgCls.Reference())
	if err != nil {
		return nil, err
	}

	return Tags(tags), err
}
