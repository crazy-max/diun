package registry

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	dockerregistry "github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupAuthDockerHub(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	auth := base64.StdEncoding.EncodeToString([]byte("janedoe:s3cr3t"))
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"auths": {
			"https://index.docker.io/v1/": {
				"auth": "`+auth+`"
			}
		}
	}`), 0o600))

	got, err := lookupAuth(configDir, "docker.io")
	require.NoError(t, err)
	assert.Equal(t, dockerregistry.AuthConfig{
		Username:      "janedoe",
		Password:      "s3cr3t",
		ServerAddress: dockerHubConfigKey,
	}, got)
}

func TestLookupAuthNotFound(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"auths":{}}`), 0o600))

	got, err := lookupAuth(configDir, "ghcr.io")
	require.NoError(t, err)
	assert.Equal(t, dockerregistry.AuthConfig{}, got)
}
