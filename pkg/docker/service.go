package docker

import (
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

// ServiceList returns Swarm services
func (c *Client) ServiceList(filterArgs filters.Args) ([]swarm.Service, error) {
	services, err := c.API.ServiceList(c.ctx, types.ServiceListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Spec.Name < services[j].Spec.Name
	})

	return services, nil
}
