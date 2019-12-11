package docker

import (
	"context"

	"github.com/docker/docker/client"
)

// Client represents an active docker object
type Client struct {
	Api *client.Client
}

// NewClient initializes a new Docker API client with default values
func NewClient(endpoint string, apiVersion string, caFile string, certFile string, keyFile string) (*Client, error) {
	d, err := client.NewClientWithOpts(
		client.WithHost(endpoint),
		client.WithVersion(apiVersion),
		client.WithTLSClientConfig(caFile, certFile, keyFile),
	)
	if err != nil {
		return nil, err
	}

	_, err = d.ServerVersion(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{Api: d}, err
}
