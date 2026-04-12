package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/regclient/regclient/internal/conffile"
	"github.com/regclient/regclient/types/errs"
)

const (
	// dockerEnv is the environment variable used to look for Docker's config.json.
	dockerEnv = "DOCKER_CONFIG"
	// dockerEnvConfig is used to inject the config as an environment variable.
	dockerEnvConfig = "DOCKER_AUTH_CONFIG"
	// dockerDir is the directory name for Docker's config (inside the users home directory).
	dockerDir = ".docker"
	// dockerConfFile is the name of Docker's config file.
	dockerConfFile = "config.json"
	// dockerHelperPre is the prefix of docker credential helpers.
	dockerHelperPre = "docker-credential-"
)

// dockerConfig is used to parse the ~/.docker/config.json
type dockerConfig struct {
	AuthConfigs       map[string]dockerAuthConfig  `json:"auths"`
	HTTPHeaders       map[string]string            `json:"HttpHeaders,omitempty"`
	DetachKeys        string                       `json:"detachKeys,omitempty"`
	CredentialsStore  string                       `json:"credsStore,omitempty"`
	CredentialHelpers map[string]string            `json:"credHelpers,omitempty"`
	Proxies           map[string]dockerProxyConfig `json:"proxies,omitempty"`
}

// dockerProxyConfig contains proxy configuration settings
type dockerProxyConfig struct {
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
	FTPProxy   string `json:"ftpProxy,omitempty"`
	AllProxy   string `json:"allProxy,omitempty"`
}

// dockerAuthConfig contains the auths
type dockerAuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"` //#nosec G117 exported struct intentionally holds secrets
	Auth     string `json:"auth,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

// DockerLoad returns a slice of hosts from the users docker config.
// This will search for the config.json in either the DOCKER_CONFIG identified directory or the default .docker directory.
// It also includes hosts extracted from the DOCKER_AUTH_CONFIG variable.
// If the config file is missing and no value is injected using an environment variable, an empty list is returned.
func DockerLoad() ([]Host, error) {
	hosts := []Host{}
	errList := []error{}
	// load from a file
	cf := conffile.New(
		conffile.WithHomeDir(dockerDir, dockerConfFile, true),
		conffile.WithEnvDir(dockerEnv, dockerConfFile),
	)
	rdr, err := cf.Open()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		errList = append(errList, err)
	} else if err == nil {
		defer rdr.Close()
		hostsFile, err := dockerParse(rdr)
		if err != nil {
			errList = append(errList, err)
		} else {
			hosts = append(hosts, hostsFile...)
		}
	}
	// load from an env var
	hostsEnv, err := DockerLoadEnv(dockerEnvConfig)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		errList = append(errList, err)
	} else if err == nil {
		hosts = append(hosts, hostsEnv...)
	}
	// return the concatenated result, only wrapping an error list if necessary
	if len(errList) == 1 {
		return hosts, errList[0]
	} else {
		return hosts, errors.Join(errList...)
	}
}

// DockerLoadFile returns a slice of hosts from a named docker config file.
func DockerLoadFile(fname string) ([]Host, error) {
	//#nosec G304 scoping file operations to a directory is not yet a feature of regclient.
	rdr, err := os.Open(fname)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return []Host{}, nil
	} else if err != nil {
		return nil, err
	}
	defer rdr.Close()
	return dockerParse(rdr)
}

// DockerLoadEnv returns a slice of hosts extracted from the config injected in an environment variable.
func DockerLoadEnv(envName string) ([]Host, error) {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return []Host{}, errs.ErrNotFound
	}
	return dockerParse(strings.NewReader(envVal))
}

// dockerParse parses a docker config into a slice of Hosts.
func dockerParse(rdr io.Reader) ([]Host, error) {
	dc := dockerConfig{}
	if err := json.NewDecoder(rdr).Decode(&dc); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	hosts := []Host{}
	for name, auth := range dc.AuthConfigs {
		if !HostValidate(name) {
			continue
		}
		h, err := dockerAuthToHost(name, dc, auth)
		if err != nil {
			continue
		}
		hosts = append(hosts, h)
	}
	// also include default entries for credential helpers
	for name, helper := range dc.CredentialHelpers {
		if !HostValidate(name) {
			continue
		}
		h := HostNewName(name)
		h.CredHelper = dockerHelperPre + helper
		if _, ok := dc.AuthConfigs[name]; ok {
			continue // skip fields with auth config
		}
		hosts = append(hosts, *h)
	}
	// add credStore entries
	if dc.CredentialsStore != "" {
		ch := newCredHelper(dockerHelperPre+dc.CredentialsStore, map[string]string{})
		csHosts, err := ch.list()
		if err == nil {
			hosts = append(hosts, csHosts...)
		}
	}
	return hosts, nil
}

// dockerAuthToHost parses an auth entry from a docker config into a Host.
func dockerAuthToHost(name string, conf dockerConfig, auth dockerAuthConfig) (Host, error) {
	helper := ""
	if conf.CredentialHelpers != nil && conf.CredentialHelpers[name] != "" {
		helper = dockerHelperPre + conf.CredentialHelpers[name]
	}
	// parse base64 auth into user/pass
	if auth.Auth != "" {
		var err error
		auth.Username, auth.Password, err = decodeAuth(auth.Auth)
		if err != nil {
			return Host{}, err
		}
	}
	if (auth.Username == "" || auth.Password == "") && auth.IdentityToken == "" && helper == "" {
		return Host{}, fmt.Errorf("no credentials found for %s", name)
	}

	h := HostNewName(name)
	// ignore unknown names
	if h.Name != DockerRegistry && !strings.HasSuffix(strings.TrimSuffix(name, "/"), h.Name) {
		return Host{}, fmt.Errorf("rejecting entry with repository: %s", name)
	}
	h.User = auth.Username
	h.Pass = auth.Password
	h.Token = auth.IdentityToken
	h.CredHelper = helper
	return *h, nil
}

// decodeAuth extracts a base64 encoded user:pass into the username and password.
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(authStr)
	if err != nil {
		return "", "", err
	}
	userPass := strings.SplitN(string(decoded), ":", 2)
	if len(userPass) != 2 {
		return "", "", fmt.Errorf("invalid auth configuration file")
	}
	return userPass[0], strings.Trim(userPass[1], "\x00"), nil
}
