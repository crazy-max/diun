package registry

import (
	"context"
	"time"

	"github.com/crazy-max/diun/v4/pkg/registry/auth"
	"github.com/pkg/errors"
	"go.podman.io/image/v5/types"
)

// Client represents an active docker registry object
type Client struct {
	opts   Options
	sysCtx *types.SystemContext
}

// Options holds docker registry object options
type Options struct {
	AuthProvider  auth.Provider
	InsecureTLS   bool
	Timeout       time.Duration
	UserAgent     string
	CompareDigest bool
	ImageOs       string
	ImageArch     string
	ImageVariant  string
}

func (c *Client) systemContext(ctx context.Context, domain string) (*types.SystemContext, error) {
	if c.opts.AuthProvider == nil {
		return c.sysCtx, nil
	}
	credentials, err := c.opts.AuthProvider.Credentials(ctx, domain)
	if err != nil {
		return nil, errors.Wrap(err, "cannot resolve registry credentials")
	}
	sys := *c.sysCtx
	sys.DockerAuthConfig = &types.DockerAuthConfig{
		Username:      credentials.Username,
		Password:      credentials.Password,
		IdentityToken: credentials.IdentityToken,
	}
	return &sys, nil
}

// New creates new docker registry client instance
func New(opts Options) (*Client, error) {
	return &Client{
		opts: opts,
		sysCtx: &types.SystemContext{
			DockerAuthConfig:                  &types.DockerAuthConfig{},
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
	var cancelFunc context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		cancelCtx, cancel := context.WithCancelCause(ctx)
		ctx, _ = context.WithTimeoutCause(cancelCtx, c.opts.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
		cancelFunc = func() { cancel(errors.WithStack(context.Canceled)) }
	}
	return ctx, cancelFunc
}
