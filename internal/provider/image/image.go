package image

import (
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider"
	"github.com/rs/zerolog/log"
)

// Client represents an active image provider object
type Client struct {
	*provider.Client
	elts []model.PrdImage
}

// New creates new image provider instance
func New(elts []model.PrdImage) *provider.Client {
	return &provider.Client{Handler: &Client{
		elts: elts,
	}}
}

// ListJob returns job list to process
func (c *Client) ListJob() []model.Job {
	if len(c.elts) == 0 {
		return []model.Job{}
	}

	log.Info().Msgf("Found %d image provider(s) to analyze...", len(c.elts))
	var list []model.Job
	for _, elt := range c.elts {
		list = append(list, model.Job{
			Provider: "image",
			ID:       elt.Name,
			Image:    model.Image(elt),
		})
	}

	return list
}
