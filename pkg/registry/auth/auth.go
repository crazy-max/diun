// Package auth provides credentials for container registries.
package auth

import "context"

// Provider resolves credentials for a registry. Implementations must support
// concurrent calls and refresh expiring credentials. Provider values must be
// comparable because registry options are used as client cache keys.
type Provider interface {
	Credentials(context.Context, string) (Credentials, error)
}

// Credentials holds registry login credentials independent of the transport.
type Credentials struct {
	Username      string
	Password      string
	IdentityToken string
}
