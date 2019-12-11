package docker

import (
	"context"

	"github.com/docker/docker/client"
)

// Client represents an active docker object
type Client struct {
	Api *client.Client
}

// NewEnvClient initializes a new Docker API client based on environment variables
func NewEnvClient() (*Client, error) {
	d, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.12"))
	if err != nil {
		return nil, err
	}

	_, err = d.ServerVersion(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{Api: d}, err
}
