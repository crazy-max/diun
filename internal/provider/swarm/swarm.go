package swarm

import (
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider"
	"github.com/rs/zerolog/log"
)

// Client represents an active swarm provider object
type Client struct {
	*provider.Client
	elts []model.PrdSwarm
}

// New creates new swarm provider instance
func New(elts []model.PrdSwarm) *provider.Client {
	return &provider.Client{Handler: &Client{
		elts: elts,
	}}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if len(c.elts) == 0 {
		return []model.Job{}
	}

	log.Info().Msgf("Found %d swarm provider(s) to analyze...", len(c.elts))
	var list []model.Job
	for _, elt := range c.elts {
		for _, img := range c.listServiceImage(elt) {
			list = append(list, model.Job{
				Provider: "swarm",
				ID:       elt.ID,
				Image:    img,
			})
		}
	}

	return list
}
