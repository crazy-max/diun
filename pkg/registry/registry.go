package registry

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"github.com/regclient/regclient"
	regconfig "github.com/regclient/regclient/config"
	regscheme "github.com/regclient/regclient/scheme/reg"
	regplatform "github.com/regclient/regclient/types/platform"
)

type Client struct {
	opts   Options
	regctl *regclient.RegClient
}

type Options struct {
	Host          *regconfig.Host
	Platform      regplatform.Platform
	Logger        *slog.Logger
	Timeout       time.Duration
	UserAgent     string
	CompareDigest bool
}

func New(opts Options) *Client {
	regctlOpts := []regclient.Opt{
		regclient.WithDockerCreds(),
		regclient.WithRegOpts(regscheme.WithDelay(2*time.Second, 60*time.Second)),
	}
	if opts.Host != nil {
		regctlOpts = append(regctlOpts, regclient.WithConfigHost(*opts.Host))
	}
	if opts.Logger != nil {
		regctlOpts = append(regctlOpts, regclient.WithSlog(opts.Logger))
	}
	if opts.UserAgent != "" {
		regctlOpts = append(regctlOpts, regclient.WithUserAgent(opts.UserAgent))
	}
	return &Client{
		opts:   opts,
		regctl: regclient.New(regctlOpts...),
	}
}

func (c *Client) timeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancelFunc context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		cancelCtx, cancel := context.WithCancelCause(ctx)
		ctx, _ = context.WithTimeoutCause(cancelCtx, c.opts.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // parent cancellation is enough
		cancelFunc = func() { cancel(errors.WithStack(context.Canceled)) }
	}
	return ctx, cancelFunc
}
