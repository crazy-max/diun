package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

// Client represents an active docker registry object
type Client struct {
	opts   Options
	sysCtx *types.SystemContext
}

// Options holds docker registry object options
type Options struct {
	Username     string
	Password     string
	InsecureTLS  bool
	Timeout      time.Duration
	UserAgent    string
	ImageOs      string
	ImageArch    string
	ImageVariant string
}

// New creates new docker registry client instance
func New(opts Options) (*Client, error) {
	// Auth
	var auth *types.DockerAuthConfig
	if opts.Username != "" {
		auth = &types.DockerAuthConfig{
			Username: opts.Username,
			Password: opts.Password,
		}
	}

	// Sys context
	sysCtx := &types.SystemContext{
		DockerAuthConfig:                  auth,
		DockerDaemonInsecureSkipTLSVerify: opts.InsecureTLS,
		DockerInsecureSkipTLSVerify:       types.NewOptionalBool(opts.InsecureTLS),
		DockerRegistryUserAgent:           opts.UserAgent,
		OSChoice:                          opts.ImageOs,
		ArchitectureChoice:                opts.ImageArch,
		VariantChoice:                     opts.ImageVariant,
	}

	return &Client{
		sysCtx: sysCtx,
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

func (c *Client) newImage(ctx context.Context, imageStr string) (types.ImageCloser, error) {
	if !strings.HasPrefix(imageStr, "//") {
		imageStr = fmt.Sprintf("//%s", imageStr)
	}

	ref, err := docker.ParseReference(imageStr)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid image name")
	}

	if c.sysCtx.DockerAuthConfig == nil {
		c.sysCtx.DockerAuthConfig = &types.DockerAuthConfig{}
		// TODO: Seek credentials
		//auth, err := config.GetCredentials(c.sysCtx, reference.Domain(ref.DockerReference()))
		//if err != nil {
		//	return nil, errors.Wrap(err, "Cannot get registry credentials")
		//}
		//*c.sysCtx.DockerAuthConfig = auth
	}

	img, err := ref.NewImage(ctx, c.sysCtx)
	if err != nil {
		return nil, err
	}

	return img, nil
}
