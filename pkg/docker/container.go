package docker

import (
	"sort"

	"github.com/moby/moby/api/types/container"
	mobyclient "github.com/moby/moby/client"
)

// ContainerList returns Docker containers
func (c *Client) ContainerList(filterArgs mobyclient.Filters) ([]container.Summary, error) {
	result, err := c.API.ContainerList(c.ctx, mobyclient.ContainerListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	containers := result.Items
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Image < containers[j].Image
	})

	return containers, nil
}
