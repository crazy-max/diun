package docker

import (
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active docker provider object
type Client struct {
	*provider.Client
	config   *model.PrdDocker
	logger   zerolog.Logger
	defaults *model.Defaults
}

// New creates new docker provider instance
func New(config *model.PrdDocker, defaults *model.Defaults) *provider.Client {
	return &provider.Client{
		Handler: NewClient(config, defaults),
	}
}

// NewClient creates a docker provider client. It's exposed separately from
// New so callers that need docker-specific behavior beyond the generic
// provider.Handler interface (currently, the Prometheus local digest
// source, see LocalDigests) can hold on to the concrete type.
func NewClient(config *model.PrdDocker, defaults *model.Defaults) *Client {
	return &Client{
		config:   config,
		logger:   log.With().Str("provider", "docker").Logger(),
		defaults: defaults,
	}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if c.config == nil {
		return []model.Job{}
	}

	images := c.listContainerImage()
	if len(images) == 0 {
		log.Warn().Msg("No image found")
		return []model.Job{}
	}

	c.logger.Info().Msgf("Found %d image(s) to analyze", len(images))
	var list []model.Job
	for _, image := range images {
		list = append(list, model.Job{
			Provider: "docker",
			Image:    image,
		})
	}

	return list
}
