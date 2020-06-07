package swarm

import (
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active swarm provider object
type Client struct {
	*provider.Client
	config *model.PrdSwarm
	logger zerolog.Logger
}

// New creates new swarm provider instance
func New(config *model.PrdSwarm) *provider.Client {
	return &provider.Client{
		Handler: &Client{
			config: config,
			logger: log.With().Str("provider", "swarm").Logger(),
		},
	}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if c.config == nil {
		return []model.Job{}
	}

	images := c.listServiceImage()
	if len(images) == 0 {
		log.Warn().Msg("No image found")
		return []model.Job{}
	}

	c.logger.Info().Msgf("Found %d image(s) to analyze", len(images))
	var list []model.Job
	for _, image := range images {
		list = append(list, model.Job{
			Provider: "swarm",
			Image:    image,
		})
	}

	return list
}
