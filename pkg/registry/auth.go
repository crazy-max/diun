package registry

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	helperclient "github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/pkg/errors"
)

// AuthConfig is comparable so identical watches can share registry clients.
type AuthConfig struct {
	Username      string
	Password      string
	IdentityToken string
}

// GetCredentials checks Docker logins before falling back to containers auth files.
func GetCredentials(domain string) (AuthConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return AuthConfig{}, err
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	dockerHome := os.Getenv("DOCKER_CONFIG")
	if dockerHome == "" {
		dockerHome = filepath.Join(home, ".docker")
	}
	paths := []string{
		filepath.Join(dockerHome, "config.json"),
		filepath.Join(home, ".dockercfg"),
		defaultAuthPath(home),
		filepath.Join(configHome, "containers", "auth.json"),
	}
	for _, path := range paths {
		auth, err := credentialsFromFile(path, domain)
		if err != nil || auth != (AuthConfig{}) {
			return auth, err
		}
	}
	return AuthConfig{}, nil
}

func credentialsFromFile(path, domain string) (AuthConfig, error) {
	// #nosec G703 -- paths come from the local user's Docker/XDG configuration, not registry input.
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return AuthConfig{}, nil
	}
	if err != nil {
		return AuthConfig{}, err
	}
	type entry struct {
		Auth          string
		IdentityToken string
	}
	var file struct {
		Auths       map[string]entry
		CredHelpers map[string]string
		CredsStore  string
	}
	if filepath.Base(path) == ".dockercfg" {
		err = json.Unmarshal(data, &file.Auths)
	} else {
		err = json.Unmarshal(data, &file)
	}
	if err != nil {
		return AuthConfig{}, errors.Wrapf(err, "reading credentials %s", path)
	}
	key := domain
	helper := file.CredHelpers[domain]
	if helper == "" && domain == "docker.io" {
		for _, alias := range []string{"https://index.docker.io/v1/", "index.docker.io", "registry-1.docker.io"} {
			if candidate := file.CredHelpers[alias]; candidate != "" {
				key, helper = alias, candidate
				break
			}
		}
	}
	if helper == "" {
		helper = file.CredsStore
		if domain == "docker.io" {
			key = "https://index.docker.io/v1/"
		}
	}
	if helper != "" {
		cred, err := helperclient.Get(helperclient.NewShellProgramFunc("docker-credential-"+helper), key)
		if credentials.IsErrCredentialsNotFound(err) {
			return AuthConfig{}, nil
		}
		if err != nil {
			return AuthConfig{}, err
		}
		if cred.Username == "<token>" {
			return AuthConfig{IdentityToken: cred.Secret}, nil
		}
		return AuthConfig{Username: cred.Username, Password: cred.Secret}, nil
	}
	value, found := file.Auths[domain]
	if !found {
		for key, candidate := range file.Auths {
			if strings.Contains(key, "://") || filepath.Base(path) == ".dockercfg" {
				key = strings.TrimPrefix(strings.TrimPrefix(key, "https://"), "http://")
				key = strings.SplitN(key, "/", 2)[0]
			}
			if key == "index.docker.io" || key == "registry-1.docker.io" {
				key = "docker.io"
			}
			if key == domain {
				value, found = candidate, true
				break
			}
		}
	}
	if !found {
		return AuthConfig{}, nil
	}
	if value.IdentityToken != "" {
		return AuthConfig{IdentityToken: value.IdentityToken}, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value.Auth)
	if err != nil {
		return AuthConfig{}, errors.Wrapf(err, "invalid credentials in %s", path)
	}
	user, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return AuthConfig{}, nil
	}
	return AuthConfig{Username: user, Password: strings.TrimRight(password, "\x00")}, nil
}
