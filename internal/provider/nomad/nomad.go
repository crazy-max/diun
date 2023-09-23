package nomad

import (
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active nomad provider object
type Client struct {
	*provider.Client
	config   *model.PrdNomad
	logger   zerolog.Logger
	defaults *model.Defaults
}

// New creates new nomad provider instance
func New(config *model.PrdNomad, defaults *model.Defaults) *provider.Client {
	return &provider.Client{
		Handler: &Client{
			config:   config,
			logger:   log.With().Str("provider", "nomad").Logger(),
			defaults: defaults,
		},
	}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if c.config == nil {
		return []model.Job{}
	}

	images := c.listTaskImages()
	if len(images) == 0 {
		log.Warn().Msg("No image found")
		return []model.Job{}
	}

	c.logger.Info().Msgf("Found %d image(s) to analyze", len(images))
	var list []model.Job
	for _, image := range images {
		list = append(list, model.Job{
			Provider: "nomad",
			Image:    image,
		})
	}

	return list
}
