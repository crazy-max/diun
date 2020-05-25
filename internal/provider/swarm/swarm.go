package swarm

import (
	"fmt"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active swarm provider object
type Client struct {
	*provider.Client
	elts   map[string]model.PrdSwarm
	logger zerolog.Logger
}

// New creates new swarm provider instance
func New(elts map[string]model.PrdSwarm) *provider.Client {
	return &provider.Client{
		Handler: &Client{
			elts:   elts,
			logger: log.With().Str("provider", "swarm").Logger(),
		},
	}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if len(c.elts) == 0 {
		return []model.Job{}
	}

	c.logger.Info().Msgf("Found %d image(s) to analyze", len(c.elts))
	var list []model.Job
	for id, elt := range c.elts {
		for _, img := range c.listServiceImage(id, elt) {
			list = append(list, model.Job{
				Provider: fmt.Sprintf("swarm-%s", id),
				Image:    img,
			})
		}
	}

	return list
}
