package registry

import (
	"context"
	"time"

	"github.com/containers/image/v5/types"
)

// Client represents an active docker registry object
type Client struct {
	opts   Options
	sysCtx *types.SystemContext
}

// Options holds docker registry object options
type Options struct {
	Auth          types.DockerAuthConfig
	InsecureTLS   bool
	Timeout       time.Duration
	UserAgent     string
	CompareDigest bool
	ImageOs       string
	ImageArch     string
	ImageVariant  string
}

// New creates new docker registry client instance
func New(opts Options) (*Client, error) {
	return &Client{
		opts: opts,
		sysCtx: &types.SystemContext{
			DockerAuthConfig:                  &opts.Auth,
			DockerDaemonInsecureSkipTLSVerify: opts.InsecureTLS,
			DockerInsecureSkipTLSVerify:       types.NewOptionalBool(opts.InsecureTLS),
			DockerRegistryUserAgent:           opts.UserAgent,
			OSChoice:                          opts.ImageOs,
			ArchitectureChoice:                opts.ImageArch,
			VariantChoice:                     opts.ImageVariant,
		},
	}, nil
}

func (c *Client) timeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.Timeout)
	}
	return ctx, cancel
}
