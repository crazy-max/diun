package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry/auth"
	"github.com/stretchr/testify/require"
)

func TestDynamicCredentials(t *testing.T) {
	p := &testAuthProvider{credentials: auth.Credentials{Username: "identity", Password: "first"}}
	c, err := New(Options{AuthProvider: p})
	require.NoError(t, err)
	first, err := c.systemContext(context.Background(), "example.azurecr.io")
	require.NoError(t, err)
	p.credentials.Password = "refreshed"
	second, err := c.systemContext(context.Background(), "example.azurecr.io")
	require.NoError(t, err)
	require.Equal(t, "first", first.DockerAuthConfig.Password)
	require.Equal(t, "refreshed", second.DockerAuthConfig.Password)
	require.NotNil(t, c.sysCtx.DockerAuthConfig)
	require.Empty(t, c.sysCtx.DockerAuthConfig)
	p.err = errors.New("identity unavailable")
	image, err := ParseImage(ParseImageOptions{Name: "example.azurecr.io/app:latest"})
	require.NoError(t, err)
	_, err = c.ListTags(image)
	require.ErrorIs(t, err, p.err)
	_, _, err = c.Manifest(image, Manifest{})
	require.ErrorIs(t, err, p.err)
}

func TestStaticCredentials(t *testing.T) {
	for _, credentials := range []auth.Static{
		{},
		{
			Username: "user",
			Password: "password",
		},
		{
			IdentityToken: "docker-identity-token",
		},
	} {
		c, err := New(Options{AuthProvider: credentials})
		require.NoError(t, err)
		sys, err := c.systemContext(context.Background(), "example.com")
		require.NoError(t, err)
		require.Equal(t, credentials.Username, sys.DockerAuthConfig.Username)
		require.Equal(t, credentials.Password, sys.DockerAuthConfig.Password)
		require.Equal(t, credentials.IdentityToken, sys.DockerAuthConfig.IdentityToken)
	}
}

func TestNoAuthProvider(t *testing.T) {
	c, err := New(Options{})
	require.NoError(t, err)
	sys, err := c.systemContext(context.Background(), "example.com")
	require.NoError(t, err)
	require.NotNil(t, sys.DockerAuthConfig)
	require.Empty(t, sys.DockerAuthConfig)
}

type testAuthProvider struct {
	credentials auth.Credentials
	err         error
}

func (p *testAuthProvider) Credentials(context.Context, string) (auth.Credentials, error) {
	return p.credentials, p.err
}
