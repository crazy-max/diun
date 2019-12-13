package docker

import (
	"context"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// Containers return containers based on filters
func (c *Client) Containers(filterArgs filters.Args) ([]types.Container, error) {
	containers, err := c.Api.ContainerList(context.Background(), types.ContainerListOptions{
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
