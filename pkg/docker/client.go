package docker

import (
	"context"
	"time"

	"github.com/containers/image/types"
)

// RegistryClient represents an active docker registry object
type RegistryClient struct {
	ctx    context.Context
	cancel context.CancelFunc
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
	// Context
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}

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
		ctx:    ctx,
		cancel: cancel,
		sysCtx: sysCtx,
	}, nil
}
