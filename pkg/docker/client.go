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

// Options holds docker client object options
type Options struct {
	Endpoint    string
	APIVersion  string
	TLSCertPath string
	TLSVerify   bool
}

// New initializes a new Docker API client with default values
func New(opts Options) (*Client, error) {
	var dockerOpts []client.Opt
	if opts.Endpoint != "" {
		dockerOpts = append(dockerOpts, client.WithHost(opts.Endpoint))
	}
	if opts.APIVersion != "" {
		dockerOpts = append(dockerOpts, client.WithVersion(opts.APIVersion))
	} else {
		dockerOpts = append(dockerOpts, client.WithAPIVersionNegotiation())
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
			return nil, errors.Wrap(err, "failed to create tls config")
		}
		httpCli := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: client.CheckRedirect,
		}
		dockerOpts = append(dockerOpts, client.WithHTTPClient(httpCli))
	}

	cli, err := client.NewClientWithOpts(dockerOpts...)
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
