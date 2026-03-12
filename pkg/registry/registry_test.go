package registry

import (
	"os"
	"testing"

	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	rc *Client
)

func TestMain(m *testing.M) {
	var err error

	rc, err = New(Options{
		ImageOs:   "linux",
		ImageArch: "amd64",
	})
	if err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	assert.NotNil(t, rc)
}

func TestNewMapsDockerRegistryAuth(t *testing.T) {
	t.Parallel()

	rc, err := New(Options{
		Auth: dockerregistry.AuthConfig{
			Username:      "janedoe",
			Password:      "s3cr3t",
			IdentityToken: "token",
		},
		InsecureTLS: true,
		UserAgent:   "diun/test",
		ImageOs:     "linux",
		ImageArch:   "amd64",
	})
	require.NoError(t, err)
	require.NotNil(t, rc.sysCtx)
	require.NotNil(t, rc.sysCtx.DockerAuthConfig)

	assert.Equal(t, "janedoe", rc.sysCtx.DockerAuthConfig.Username)
	assert.Equal(t, "s3cr3t", rc.sysCtx.DockerAuthConfig.Password)
	assert.Equal(t, "token", rc.sysCtx.DockerAuthConfig.IdentityToken)
	assert.Equal(t, "diun/test", rc.sysCtx.DockerRegistryUserAgent)
	assert.Equal(t, "linux", rc.sysCtx.OSChoice)
	assert.Equal(t, "amd64", rc.sysCtx.ArchitectureChoice)
}
