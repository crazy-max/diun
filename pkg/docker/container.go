package docker

import (
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// ContainerList returns Docker containers
func (c *Client) ContainerList(filterArgs filters.Args) ([]types.Container, error) {
	containers, err := c.Api.ContainerList(c.ctx, types.ContainerListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Image < containers[j].Image
	})

	return containers, nil
}
