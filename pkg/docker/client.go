package docker

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/docker/go-connections/tlsconfig"
	mobyclient "github.com/moby/moby/client"
	"github.com/pkg/errors"
)

// Client represents an active docker object
type Client struct {
	ctx context.Context
	API *mobyclient.Client
}

// Options holds docker client object options
type Options struct {
	Endpoint    string
	APIVersion  string
	TLSCertPath string
	TLSVerify   bool
}

// New initializes a new Docker API client with default values
func New(opts Options) (*Client, error) {
	var dockerOpts []mobyclient.Opt
	if opts.Endpoint != "" {
		dockerOpts = append(dockerOpts, mobyclient.WithHost(opts.Endpoint))
	}
	if opts.APIVersion != "" {
		dockerOpts = append(dockerOpts, mobyclient.WithAPIVersion(opts.APIVersion))
	}
	if opts.TLSCertPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(opts.TLSCertPath, "ca.pem"),
			CertFile:           filepath.Join(opts.TLSCertPath, "cert.pem"),
			KeyFile:            filepath.Join(opts.TLSCertPath, "key.pem"),
			InsecureSkipVerify: !opts.TLSVerify,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create TLS config")
		}
		httpCli := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: mobyclient.CheckRedirect,
		}
		dockerOpts = append(dockerOpts, mobyclient.WithHTTPClient(httpCli))
	}

	cli, err := mobyclient.New(dockerOpts...)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	_, err = cli.ServerVersion(ctx, mobyclient.ServerVersionOptions{})
	if err != nil {
		return nil, err
	}

	return &Client{
		ctx: ctx,
		API: cli,
	}, err
}

// Close closes docker client
func (c *Client) Close() {
	if c.API != nil {
		_ = c.API.Close()
	}
}
