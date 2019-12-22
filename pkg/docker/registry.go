package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containers/image/docker"
	"github.com/containers/image/types"
)

// RegistryClient represents an active docker registry object
type RegistryClient struct {
	opts   RegistryOptions
	sysCtx *types.SystemContext
}

// RegistryOptions holds docker registry object options
type RegistryOptions struct {
	Os          string
	Arch        string
	Username    string
	Password    string
	InsecureTLS bool
	Timeout     time.Duration
}

// NewRegistryClient creates new docker registry client instance
func NewRegistryClient(opts RegistryOptions) (*RegistryClient, error) {
	// Auth
	auth := &types.DockerAuthConfig{}
	if opts.Username != "" {
		auth = &types.DockerAuthConfig{
			Username: opts.Username,
			Password: opts.Password,
		}
	}

	// Sys context
	sysCtx := &types.SystemContext{
		OSChoice:                          opts.Os,
		ArchitectureChoice:                opts.Arch,
		DockerAuthConfig:                  auth,
		DockerDaemonInsecureSkipTLSVerify: opts.InsecureTLS,
		DockerInsecureSkipTLSVerify:       types.NewOptionalBool(opts.InsecureTLS),
	}

	return &RegistryClient{
		sysCtx: sysCtx,
	}, nil
}

func (c *RegistryClient) timeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.Timeout)
	}
	return ctx, cancel
}

func (c *RegistryClient) newImage(ctx context.Context, imageStr string) (types.ImageCloser, error) {
	if !strings.HasPrefix(imageStr, "//") {
		imageStr = fmt.Sprintf("//%s", imageStr)
	}

	ref, err := docker.ParseReference(imageStr)
	if err != nil {
		return nil, fmt.Errorf("invalid image name %s: %v", imageStr, err)
	}

	img, err := ref.NewImage(ctx, c.sysCtx)
	if err != nil {
		return nil, err
	}

	return img, nil
}
