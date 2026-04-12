package registry

import (
	"testing"

	regconfig "github.com/regclient/regclient/config"
	regplatform "github.com/regclient/regclient/types/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client := New(Options{
		Host: regconfig.HostNewName("docker.io"),
		Platform: regplatform.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	})
	assert.NotNil(t, client)
}

func TestNewMapsRegistryAuth(t *testing.T) {
	client := New(Options{
		Host: &regconfig.Host{
			Name:  "docker.io",
			User:  "janedoe",
			Pass:  "s3cr3t",
			Token: "token",
			TLS:   regconfig.TLSInsecure,
		},
		UserAgent: "diun/test",
		Platform: regplatform.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
	})
	require.NotNil(t, client.opts.Host)
	require.Equal(t, "janedoe", client.opts.Host.User)
	require.Equal(t, "s3cr3t", client.opts.Host.Pass)
	require.Equal(t, "token", client.opts.Host.Token)
	require.Equal(t, "diun/test", client.opts.UserAgent)
	require.Equal(t, regplatform.Platform{
		OS:           "linux",
		Architecture: "amd64",
	}, client.opts.Platform)
}

func TestNewUsesLocalPlatformByDefault(t *testing.T) {
	client := New(Options{
		Platform: regplatform.Local(),
	})
	require.Nil(t, client.opts.Host)
	require.Equal(t, regplatform.Local(), client.opts.Platform)
}
