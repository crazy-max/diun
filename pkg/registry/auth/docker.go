package auth

import (
	"context"

	"go.podman.io/image/v5/pkg/docker/config"
)

// Docker supplies discovered Docker credentials. Its value is comparable so
// Docker providers with identical credentials can share registry clients.
type Docker struct {
	credentials Credentials
}

// NewDocker discovers credentials using Docker config files and credential
// helpers. Discovery happens once per job, as with explicit static credentials.
func NewDocker(domain string) (Docker, error) {
	credentials, err := config.GetCredentials(nil, domain)
	if err != nil {
		return Docker{}, err
	}
	return Docker{
		credentials: Credentials{
			Username:      credentials.Username,
			Password:      credentials.Password,
			IdentityToken: credentials.IdentityToken,
		},
	}, nil
}

func (d Docker) Credentials(context.Context, string) (Credentials, error) {
	return d.credentials, nil
}
