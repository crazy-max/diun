package registry

import (
	"github.com/containers/image/docker"
)

type Tags []string

// Tags returns tags of a Docker repository
func (c *Client) Tags(opts *Options) (Tags, error) {
	ctx, cancel := c.timeoutContext(opts.Timeout)
	defer cancel()

	img, sys, err := c.newImage(ctx, opts)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	tags, err := docker.GetRepositoryTags(ctx, sys, img.Reference())
	if err != nil {
		return nil, err
	}

	return Tags(tags), err
}
