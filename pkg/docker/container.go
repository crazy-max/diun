package docker

import (
	"context"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// ContainerOptions holds docker container object options
type ContainerOptions struct {
	IncludeStopped bool
}

// ContainerList return container list.
func (c *Client) ContainerList(filterArgs ...filters.KeyValuePair) ([]types.Container, error) {
	containers, err := c.Api.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(filterArgs...),
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Image < containers[j].Image
	})

	return containers, nil
}
