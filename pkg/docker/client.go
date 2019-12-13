package docker

import (
	"context"

	"github.com/docker/docker/client"
)

// Client represents an active docker object
type Client struct {
	context context.Context
	Api     *client.Client
}

// NewClient initializes a new Docker API client with default values
func NewClient(endpoint string, apiVersion string, caFile string, certFile string, keyFile string) (*Client, error) {
	var opts []client.Opt
	if endpoint != "" {
		opts = append(opts, client.WithHost(endpoint))
	}
	if apiVersion != "" {
		opts = append(opts, client.WithVersion(apiVersion))
	}
	if caFile != "" && certFile != "" && keyFile != "" {
		opts = append(opts, client.WithTLSClientConfig(caFile, certFile, keyFile))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	_, err = cli.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		context: ctx,
		Api:     cli,
	}, err
}
