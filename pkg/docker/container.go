package docker

import (
	"sort"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// ContainerList returns Docker containers
func (c *Client) ContainerList(filterArgs filters.Args) ([]container.Summary, error) {
	containers, err := c.API.ContainerList(c.ctx, container.ListOptions{
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
