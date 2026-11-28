package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/crazy-max/diun/v4/internal/logging"
	"github.com/pkg/errors"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/scheme/reg"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"golang.org/x/sync/singleflight"
)

// Client represents an active docker registry object
type Client struct {
	opts     Options
	mu       sync.Mutex
	init     singleflight.Group
	clients  map[string]*registryClients
	platform platform.Platform
}

// HEAD checks have their own backoff so exhausted pull quotas do not delay
// checks of unchanged images. Downloads still share auth, cache and retries.
type registryClients struct {
	head    *regclient.RegClient
	content *regclient.RegClient
}

// Options holds docker registry object options
type Options struct {
	Auth          AuthConfig
	InsecureTLS   bool
	Timeout       time.Duration
	UserAgent     string
	CompareDigest bool
	Platform      platform.Platform
}

// New creates new docker registry client instance
func New(opts Options) (*Client, error) {
	p := opts.Platform
	local := platform.Local()
	if p.OS == "" {
		p.OS = local.OS
	}
	if p.Architecture == "" {
		p.Architecture = local.Architecture
		if p.Variant == "" {
			p.Variant = local.Variant
		}
	}
	p.OSFeatures = slices.Clone(p.OSFeatures)
	p.Features = slices.Clone(p.Features)
	return &Client{
		opts:     opts,
		platform: p,
		clients:  make(map[string]*registryClients),
	}, nil
}

// registryClient shares auth, immutable manifest caching and backoff within a scan.
func (c *Client) registryClient(ctx context.Context, domain string) (*registryClients, error) {
	c.mu.Lock()
	client := c.clients[domain]
	c.mu.Unlock()
	if client != nil {
		return client, nil
	}
	result := c.init.DoChan(domain, func() (any, error) {
		c.mu.Lock()
		client := c.clients[domain]
		c.mu.Unlock()
		if client != nil {
			return client, nil
		}
		client, err := c.newRegistryClients(ctx, domain)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.clients[domain] = client
		c.mu.Unlock()
		return client, nil
	})
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case result := <-result:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*registryClients), nil
	}
}

func (c *Client) newRegistryClients(ctx context.Context, domain string) (*registryClients, error) {
	host := config.HostNewName(domain)
	host.User, host.Pass, host.Token = c.opts.Auth.Username, c.opts.Auth.Password, c.opts.Auth.IdentityToken
	certDirs := []string{"/etc/containers/certs.d", regclient.DockerCertDir}
	if home, err := os.UserHomeDir(); err == nil {
		certDirs = append([]string{filepath.Join(home, ".config", "containers", "certs.d")}, certDirs...)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	// Our pagination wrapper hides the concrete transport from regclient, so
	// configure TLS here rather than relying on its transport type assertion.
	transport.TLSClientConfig.InsecureSkipVerify = c.opts.InsecureTLS
	if !c.opts.InsecureTLS {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "cannot load system certificates")
		}
		for _, dir := range certDirs {
			entries, err := os.ReadDir(filepath.Join(dir, host.Hostname))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, errors.Wrap(err, "cannot read registry certificates")
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".crt") {
					continue
				}
				pem, err := os.ReadFile(filepath.Join(dir, host.Hostname, entry.Name()))
				if err != nil {
					return nil, errors.Wrap(err, "cannot read registry CA")
				}
				if !pool.AppendCertsFromPEM(pem) {
					return nil, errors.New("cannot parse registry CA")
				}
			}
		}
		transport.TLSClientConfig.RootCAs = pool
	}
	for _, dir := range certDirs {
		dir = filepath.Join(dir, domain)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot read registry certificates")
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".cert") {
				continue
			}
			cert, err := tls.LoadX509KeyPair(filepath.Join(dir, entry.Name()), filepath.Join(dir, strings.TrimSuffix(entry.Name(), ".cert")+".key"))
			if err != nil {
				return nil, errors.Wrap(err, "cannot load registry client certificate")
			}
			transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
		}
		break
	}

	newClient := func(extra ...reg.Opts) *regclient.RegClient {
		regOpts := []reg.Opts{
			reg.WithHTTPClient(&http.Client{
				Transport: registryTransport{
					base: transport.Clone(),
				},
			}),
			reg.WithCache(5*time.Minute, 1024),
		}
		return regclient.New(
			regclient.WithUserAgent(c.opts.UserAgent),
			regclient.WithSlog(slog.New(logging.SlogHandler{})),
			regclient.WithConfigHost(*host),
			regclient.WithRegOpts(append(regOpts, extra...)...),
		)
	}
	if c.opts.InsecureTLS {
		host.TLS = config.TLSInsecure
		// Probe with the same certificates but a separate retry budget. Only an
		// explicit HTTP response at the registry permits falling back to HTTP.
		probe := newClient(reg.WithRetryLimit(1))
		_, err := probe.Ping(ctx, ref.Ref{Scheme: "reg", Registry: domain})
		if context.Cause(ctx) != nil {
			return nil, context.Cause(ctx)
		}
		if transportErr, ok := stderrors.AsType[*url.Error](err); ok && stderrors.Is(err, http.ErrSchemeMismatch) {
			endpoint, parseErr := url.Parse(transportErr.URL)
			if parseErr == nil && endpoint.Scheme == "https" && endpoint.Host == host.Hostname {
				host.TLS = config.TLSDisabled
				probe = newClient(reg.WithRetryLimit(1))
				_, err = probe.Ping(ctx, ref.Ref{Scheme: "reg", Registry: domain})
			}
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot probe registry")
		}
	}
	client := &registryClients{
		head:    newClient(),
		content: newClient(),
	}
	return client, nil
}

func (c *Client) timeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancelFunc context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		cancelCtx, cancel := context.WithCancelCause(ctx)
		ctx, _ = context.WithTimeoutCause(cancelCtx, c.opts.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
		cancelFunc = func() { cancel(errors.WithStack(context.Canceled)) }
	}
	return ctx, cancelFunc
}
