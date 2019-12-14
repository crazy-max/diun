package docker

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
)

// Client represents an active docker object
type Client struct {
	ctx context.Context
	API *client.Client
}

// NewClient initializes a new Docker API client with default values
func NewClient(endpoint, apiVersion, tlsCertsPath string, tlsVerify bool) (*Client, error) {
	var opts []client.Opt
	if endpoint != "" {
		opts = append(opts, client.WithHost(endpoint))
	}
	if apiVersion != "" {
		opts = append(opts, client.WithVersion(apiVersion))
	}
	if tlsCertsPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(tlsCertsPath, "ca.pem"),
			CertFile:           filepath.Join(tlsCertsPath, "cert.pem"),
			KeyFile:            filepath.Join(tlsCertsPath, "key.pem"),
			InsecureSkipVerify: !tlsVerify,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tls config")
		}
		httpCli := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: client.CheckRedirect,
		}
		opts = append(opts, client.WithHTTPClient(httpCli))
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
		ctx: ctx,
		API: cli,
	}, err
}
