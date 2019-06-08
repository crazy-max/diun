package docker

import (
	"github.com/containers/image/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
)

type Tags []string

// Tags returns tags of a Docker repository
func (c *RegistryClient) Tags(image registry.Image, max int) (Tags, int, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgCls, err := c.newImage(ctx, image.String())
	if err != nil {
		return nil, 0, err
	}
	defer imgCls.Close()

	tags, err := docker.GetRepositoryTags(ctx, c.sysCtx, imgCls.Reference())
	if err != nil {
		return nil, 0, err
	}

	// Reverse order (latest tags first)
	for i := len(tags)/2 - 1; i >= 0; i-- {
		opp := len(tags) - 1 - i
		tags[i], tags[opp] = tags[opp], tags[i]
	}

	if max > 0 && len(tags) >= max {
		return Tags(tags[:max]), len(tags), nil
	}

	return Tags(tags), len(tags), nil
}
