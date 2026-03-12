package registry

import (
	"io"

	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	dockerregistry "github.com/docker/docker/api/types/registry"
)

const dockerHubConfigKey = "https://index.docker.io/v1/"

// LookupAuth returns Docker registry credentials for the given registry domain.
// If no credentials are configured, an empty AuthConfig is returned.
func LookupAuth(domain string) (dockerregistry.AuthConfig, error) {
	return lookupAuth("", domain)
}

func lookupAuth(configDir, domain string) (dockerregistry.AuthConfig, error) {
	cfg, err := loadDockerConfig(configDir)
	if err != nil {
		return dockerregistry.AuthConfig{}, err
	}

	auth, err := cfg.GetAuthConfig(dockerConfigKey(domain))
	if err != nil {
		return dockerregistry.AuthConfig{}, err
	}

	return dockerregistry.AuthConfig{
		Username:      auth.Username,
		Password:      auth.Password,
		Auth:          auth.Auth,
		ServerAddress: auth.ServerAddress,
		IdentityToken: auth.IdentityToken,
		RegistryToken: auth.RegistryToken,
	}, nil
}

func loadDockerConfig(configDir string) (*configfile.ConfigFile, error) {
	if configDir == "" {
		return dockerconfig.LoadDefaultConfigFile(io.Discard), nil
	}
	return dockerconfig.Load(configDir)
}

func dockerConfigKey(domain string) string {
	if domain == "docker.io" || domain == "index.docker.io" {
		return dockerHubConfigKey
	}
	return domain
}
