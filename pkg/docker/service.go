package docker

import (
	"sort"

	"github.com/moby/moby/api/types/swarm"
	mobyclient "github.com/moby/moby/client"
)

// ServiceList returns Swarm services
func (c *Client) ServiceList(filterArgs mobyclient.Filters) ([]swarm.Service, error) {
	result, err := c.API.ServiceList(c.ctx, mobyclient.ServiceListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	services := result.Items
	sort.Slice(services, func(i, j int) bool {
		return services[i].Spec.Name < services[j].Spec.Name
	})

	return services, nil
}
