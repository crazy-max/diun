package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containers/image/docker"
	"github.com/containers/image/types"
)

// Client represents an active registry object
type Client struct{}

type Options struct {
	Image       Image
	Username    string
	Password    string
	InsecureTLS bool
	Timeout     time.Duration
}

// New creates new registry instance
func New() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) timeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	}
	return ctx, cancel
}

func (c *Client) newImage(ctx context.Context, opts *Options) (types.ImageCloser, *types.SystemContext, error) {
	image := opts.Image.String()
	if !strings.HasPrefix(opts.Image.String(), "//") {
		image = fmt.Sprintf("//%s", opts.Image.String())
	}

	ref, err := docker.ParseReference(image)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid image name %s: %v", image, err)
	}

	auth := &types.DockerAuthConfig{}
	if opts.Username != "" {
		auth = &types.DockerAuthConfig{
			Username: opts.Username,
			Password: opts.Password,
		}
	}

	sys := &types.SystemContext{
		DockerAuthConfig:                  auth,
		DockerDaemonInsecureSkipTLSVerify: opts.InsecureTLS,
		DockerInsecureSkipTLSVerify:       types.NewOptionalBool(opts.InsecureTLS),
	}

	img, err := ref.NewImage(ctx, sys)
	if err != nil {
		return nil, nil, err
	}

	return img, sys, nil
}
