// Package regclient is used to access OCI registries.
package regclient

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/internal/version"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/scheme/ocidir"
	"github.com/regclient/regclient/scheme/reg"
)

const (
	// DefaultUserAgent sets the header on http requests.
	DefaultUserAgent = "regclient/regclient"
	// DockerCertDir default location for docker certs.
	DockerCertDir = "/etc/docker/certs.d"
	// DockerRegistry is the well known name of Docker Hub, "docker.io".
	DockerRegistry = config.DockerRegistry
	// DockerRegistryAuth is the name of Docker Hub seen in docker's config.json.
	DockerRegistryAuth = config.DockerRegistryAuth
	// DockerRegistryDNS is the actual registry DNS name for Docker Hub.
	DockerRegistryDNS = config.DockerRegistryDNS
)

// RegClient is used to access OCI distribution-spec registries.
type RegClient struct {
	hosts       map[string]*config.Host
	hostDefault *config.Host
	regOpts     []reg.Opts
	schemes     map[string]scheme.API
	slog        *slog.Logger
	userAgent   string
}

// Opt functions are used by [New] to create a [*RegClient].
type Opt func(*RegClient)

// New returns a registry client.
func New(opts ...Opt) *RegClient {
	rc := RegClient{
		hosts:     map[string]*config.Host{},
		userAgent: DefaultUserAgent,
		regOpts:   []reg.Opts{},
		schemes:   map[string]scheme.API{},
		slog:      slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}

	info := version.GetInfo()
	if info.VCSTag != "" {
		rc.userAgent = fmt.Sprintf("%s (%s)", rc.userAgent, info.VCSTag)
	} else {
		rc.userAgent = fmt.Sprintf("%s (%s)", rc.userAgent, info.VCSRef)
	}

	// inject Docker Hub settings
	_ = rc.hostSet(*config.HostNewName(config.DockerRegistryAuth))

	for _, opt := range opts {
		opt(&rc)
	}

	// configure regOpts
	hostList := []*config.Host{}
	for _, h := range rc.hosts {
		hostList = append(hostList, h)
	}
	rc.regOpts = append(rc.regOpts,
		reg.WithConfigHosts(hostList),
		reg.WithConfigHostDefault(rc.hostDefault),
		reg.WithSlog(rc.slog),
		reg.WithUserAgent(rc.userAgent),
	)

	// setup scheme's
	rc.schemes["reg"] = reg.New(rc.regOpts...)
	rc.schemes["ocidir"] = ocidir.New(
		ocidir.WithSlog(rc.slog),
	)

	rc.slog.Debug("regclient initialized",
		slog.String("VCSRef", info.VCSRef),
		slog.String("VCSTag", info.VCSTag))

	return &rc
}

// WithBlobLimit sets the max size for chunked blob uploads which get stored in memory.
//
// Deprecated: replace with WithRegOpts(reg.WithBlobLimit(limit)), see [WithRegOpts] and [reg.WithBlobLimit].
//
//go:fix inline
func WithBlobLimit(limit int64) Opt {
	return WithRegOpts(reg.WithBlobLimit(limit))
}

// WithBlobSize overrides default blob sizes.
//
// Deprecated: replace with WithRegOpts(reg.WithBlobSize(chunk, max)), see [WithRegOpts] and [reg.WithBlobSize].
//
//go:fix inline
func WithBlobSize(chunk, max int64) Opt {
	return WithRegOpts(reg.WithBlobSize(chunk, max))
}

// WithCertDir adds a path of certificates to trust similar to Docker's /etc/docker/certs.d.
//
// Deprecated: replace with WithRegOpts(reg.WithCertDirs(path)), see [WithRegOpts] and [reg.WithCertDirs].
//
//go:fix inline
func WithCertDir(path ...string) Opt {
	return WithRegOpts(reg.WithCertDirs(path))
}

// WithConfigHost adds a list of config host settings.
func WithConfigHost(configHost ...config.Host) Opt {
	return func(rc *RegClient) {
		rc.hostLoad("host", configHost)
	}
}

// WithConfigHostDefault adds default settings for new hosts.
func WithConfigHostDefault(configHost config.Host) Opt {
	return func(rc *RegClient) {
		rc.hostDefault = &configHost
	}
}

// WithConfigHosts adds a list of config host settings.
//
// Deprecated: replace with [WithConfigHost].
//
//go:fix inline
func WithConfigHosts(configHosts []config.Host) Opt {
	return WithConfigHost(configHosts...)
}

// WithDockerCerts adds certificates trusted by docker in /etc/docker/certs.d.
func WithDockerCerts() Opt {
	return WithRegOpts(reg.WithCertDirs([]string{DockerCertDir}))
}

// WithDockerCreds adds configuration from users docker config with registry logins.
// This changes the default value from the config file, and should be added after the config file is loaded.
func WithDockerCreds() Opt {
	return func(rc *RegClient) {
		configHosts, err := config.DockerLoad()
		if err != nil {
			rc.slog.Warn("Failed to load docker creds",
				slog.String("err", err.Error()))
			return
		}
		rc.hostLoad("docker", configHosts)
	}
}

// WithDockerCredsFile adds configuration from a named docker config file with registry logins.
// This changes the default value from the config file, and should be added after the config file is loaded.
func WithDockerCredsFile(fname string) Opt {
	return func(rc *RegClient) {
		configHosts, err := config.DockerLoadFile(fname)
		if err != nil {
			rc.slog.Warn("Failed to load docker creds",
				slog.String("err", err.Error()))
			return
		}
		rc.hostLoad("docker-file", configHosts)
	}
}

// WithRegOpts passes through opts to the reg scheme.
func WithRegOpts(opts ...reg.Opts) Opt {
	return func(rc *RegClient) {
		if len(opts) == 0 {
			return
		}
		rc.regOpts = append(rc.regOpts, opts...)
	}
}

// WithRetryDelay specifies the time permitted for retry delays.
//
// Deprecated: replace with WithRegOpts(reg.WithDelay(delayInit, delayMax)), see [WithRegOpts] and [reg.WithDelay].
//
//go:fix inline
func WithRetryDelay(delayInit, delayMax time.Duration) Opt {
	return WithRegOpts(reg.WithDelay(delayInit, delayMax))
}

// WithRetryLimit specifies the number of retries for non-fatal errors.
//
// Deprecated: replace with WithRegOpts(reg.WithRetryLimit(retryLimit)), see [WithRegOpts] and [reg.WithRetryLimit].
//
//go:fix inline
func WithRetryLimit(retryLimit int) Opt {
	return WithRegOpts(reg.WithRetryLimit(retryLimit))
}

// WithSlog configures the slog Logger.
func WithSlog(slog *slog.Logger) Opt {
	return func(rc *RegClient) {
		rc.slog = slog
	}
}

// WithUserAgent specifies the User-Agent http header.
func WithUserAgent(ua string) Opt {
	return func(rc *RegClient) {
		rc.userAgent = ua
	}
}

func (rc *RegClient) hostLoad(src string, hosts []config.Host) {
	for _, configHost := range hosts {
		if configHost.Name == "" {
			if configHost.Pass != "" {
				configHost.Pass = "***"
			}
			if configHost.Token != "" {
				configHost.Token = "***"
			}
			rc.slog.Warn("Ignoring registry config without a name",
				slog.Any("entry", configHost))
			continue
		}
		if configHost.Name == DockerRegistry || configHost.Name == DockerRegistryDNS || configHost.Name == DockerRegistryAuth {
			configHost.Name = DockerRegistry
			if configHost.Hostname == "" || configHost.Hostname == DockerRegistry || configHost.Hostname == DockerRegistryAuth {
				configHost.Hostname = DockerRegistryDNS
			}
		}
		tls, _ := configHost.TLS.MarshalText()
		rc.slog.Debug("Loading config",
			slog.Int64("blobChunk", configHost.BlobChunk),
			slog.Int64("blobMax", configHost.BlobMax),
			slog.String("helper", configHost.CredHelper),
			slog.String("hostname", configHost.Hostname),
			slog.Any("mirrors", configHost.Mirrors),
			slog.String("name", configHost.Name),
			slog.String("pathPrefix", configHost.PathPrefix),
			slog.Bool("repoAuth", configHost.RepoAuth),
			slog.String("source", src),
			slog.String("tls", string(tls)),
			slog.String("user", configHost.User))
		err := rc.hostSet(configHost)
		if err != nil {
			rc.slog.Warn("Failed to update host config",
				slog.String("host", configHost.Name),
				slog.String("user", configHost.User),
				slog.String("error", err.Error()))
		}
	}
}

func (rc *RegClient) hostSet(newHost config.Host) error {
	name := newHost.Name
	var err error
	if _, ok := rc.hosts[name]; !ok {
		// merge newHost with default host settings
		rc.hosts[name] = config.HostNewDefName(rc.hostDefault, name)
		err = rc.hosts[name].Merge(newHost, nil)
	} else {
		// merge newHost with existing settings
		err = rc.hosts[name].Merge(newHost, rc.slog)
	}
	if err != nil {
		return err
	}
	return nil
}
