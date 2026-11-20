package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/regclient/regclient/types/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	rc *Client
)

func TestMain(m *testing.M) {
	var err error

	rc, err = New(Options{
		Platform: platform.Platform{OS: "linux", Architecture: "amd64"},
	})
	if err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	assert.NotNil(t, rc)
}

func TestNewPlatform(t *testing.T) {
	local := platform.Local()
	for _, tt := range []struct {
		name string
		in   platform.Platform
		want platform.Platform
	}{
		{
			name: "default",
			want: platform.Platform{OS: local.OS, Architecture: local.Architecture, Variant: local.Variant},
		},
		{
			name: "OS override",
			in:   platform.Platform{OS: "linux"},
			want: platform.Platform{OS: "linux", Architecture: local.Architecture, Variant: local.Variant},
		},
		{
			name: "explicit architecture does not inherit host variant",
			in:   platform.Platform{Architecture: "arm"},
			want: platform.Platform{OS: local.OS, Architecture: "arm"},
		},
		{
			name: "variant override",
			in:   platform.Platform{Variant: "v7"},
			want: platform.Platform{OS: local.OS, Architecture: local.Architecture, Variant: "v7"},
		},
		{
			name: "full platform",
			in: platform.Platform{
				OS:           "windows",
				Architecture: "amd64",
				Variant:      "v2",
				OSVersion:    "10.0.20348.0",
				OSFeatures:   []string{"win32k"},
				Features:     []string{"sse4"},
			},
			want: platform.Platform{
				OS:           "windows",
				Architecture: "amd64",
				Variant:      "v2",
				OSVersion:    "10.0.20348.0",
				OSFeatures:   []string{"win32k"},
				Features:     []string{"sse4"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(Options{Platform: tt.in})
			require.NoError(t, err)
			if len(tt.in.OSFeatures) > 0 {
				tt.in.OSFeatures[0] = "changed"
			}
			if len(tt.in.Features) > 0 {
				tt.in.Features[0] = "changed"
			}
			assert.Equal(t, tt.want, client.platform)
		})
	}
}

func TestRegistryTransportAndAuth(t *testing.T) {
	for _, tt := range []struct {
		name      string
		tls       bool
		insecure  bool
		wantError bool
	}{
		{
			name:     "HTTP allowed",
			insecure: true,
		},
		{
			name:     "self-signed HTTPS allowed",
			tls:      true,
			insecure: true,
		},
		{
			name:      "self-signed HTTPS rejected",
			tls:       true,
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/" {
					w.WriteHeader(http.StatusOK)
					return
				}
				assert.Equal(t, "diun-test", r.UserAgent())
				user, pass, ok := r.BasicAuth()
				if !ok || user != "user" || pass != "secret" {
					w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"name":"test","tags":["latest"]}`))
			})
			var server *httptest.Server
			if tt.tls {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			t.Cleanup(server.Close)
			image, err := ParseImage(ParseImageOptions{Name: strings.TrimPrefix(strings.TrimPrefix(server.URL, "https://"), "http://") + "/test:latest"})
			require.NoError(t, err)
			client, err := New(Options{InsecureTLS: tt.insecure, Timeout: time.Second, UserAgent: "diun-test", Auth: AuthConfig{Username: "user", Password: "secret"}})
			require.NoError(t, err)
			tags, err := client.ListTags(image)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{"latest"}, tags)
		})
	}
}

func TestRegistryCanceledProbe(t *testing.T) {
	client, err := New(Options{InsecureTLS: true})
	require.NoError(t, err)
	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(context.Canceled)
	_, err = client.registryClient(ctx, "localhost:1")
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, client.clients)
}

func TestRegistryProbeDoesNotBlockOtherDomains(t *testing.T) {
	entered := make(chan struct{})
	var once sync.Once
	slow := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(entered) })
		<-r.Context().Done()
	}))
	t.Cleanup(slow.Close)
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(context.Canceled) })
	client := newTestRegistryClient(t, Options{})
	slowResult := make(chan error, 1)
	slowHost := strings.TrimPrefix(slow.URL, "https://")
	go func() {
		_, err := client.registryClient(ctx, slowHost)
		slowResult <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("slow registry probe never started")
	}
	ready := newTestRegistry(t, "test")
	readyCtx, readyCancel := context.WithTimeoutCause(t.Context(), time.Second, context.DeadlineExceeded)
	defer readyCancel()
	_, err := client.registryClient(readyCtx, ready.host())
	require.NoError(t, err, "another domain must not wait for a stalled probe")
	waitCtx, waitCancel := context.WithTimeoutCause(t.Context(), 100*time.Millisecond, context.DeadlineExceeded)
	defer waitCancel()
	_, err = client.registryClient(waitCtx, slowHost)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	cancel(context.Canceled)
	require.ErrorIs(t, <-slowResult, context.Canceled)
}

func TestRegistryConcurrentClientInitialization(t *testing.T) {
	server := newTestRegistry(t, "test")
	server.addTagsPage("", []string{"latest"}, "")
	server.addImage("latest", newTestRegistryImage(t, ocispecs.Platform{OS: "linux", Architecture: "amd64"}, "", nil))
	r, err := ImageReference(server.imageName("latest"))
	require.NoError(t, err)
	for range 10 {
		client := newTestRegistryClient(t, Options{})
		clients, err := client.registryClient(t.Context(), server.host())
		require.NoError(t, err)
		start := make(chan struct{})
		results := make(chan error, 2)
		go func() {
			<-start
			_, err := clients.head.ManifestHead(t.Context(), r)
			results <- err
		}()
		go func() {
			<-start
			_, err := clients.content.TagList(t.Context(), r)
			results <- err
		}()
		close(start)
		require.NoError(t, <-results)
		require.NoError(t, <-results)
	}
}

func TestRegistryFailedHTTPProbeIsNotCached(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	client := newTestRegistryClient(t, Options{Timeout: time.Second})
	_, err := client.registryClient(t.Context(), strings.TrimPrefix(server.URL, "http://"))
	require.Error(t, err)
	assert.Empty(t, client.clients)
}

func TestRegistryHeadRateLimitStillBacksOff(t *testing.T) {
	var heads atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			return
		}
		heads.Add(1)
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)
	client := newTestRegistryClient(t, Options{Timeout: 100 * time.Millisecond})
	image, err := ParseImage(ParseImageOptions{Name: strings.TrimPrefix(server.URL, "https://") + "/test:latest"})
	require.NoError(t, err)
	for range 2 {
		_, _, err := client.Manifest(image, Manifest{})
		require.Error(t, err)
	}
	assert.Equal(t, int32(1), heads.Load(), "HEAD throttling must not be bypassed by recreating clients")
}

func TestRegistryTLSWithClientCertificate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"test","tags":["latest"]}`))
	}))
	server.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	t.Cleanup(server.Close)
	host := strings.TrimPrefix(server.URL, "https://")
	dir := filepath.Join(home, ".config", "containers", "certs.d", host)
	if strings.Contains(host, ":") && os.PathSeparator == '\\' {
		t.Skip("Windows paths cannot contain the registry port; covered by Linux tests")
	}
	require.NoError(t, os.MkdirAll(dir, 0o700))
	cert := server.TLS.Certificates[0]
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "client.cert"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "client.key"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0o600))
	image, err := ParseImage(ParseImageOptions{Name: host + "/test:latest"})
	require.NoError(t, err)
	for _, insecure := range []bool{false, true} {
		client, err := New(Options{InsecureTLS: insecure, Timeout: time.Second})
		require.NoError(t, err)
		tags, err := client.ListTags(image)
		require.NoError(t, err)
		assert.Equal(t, []string{"latest"}, tags)
	}
}

func TestManifestHeadFailureDoesNotFetchManifest(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
	}{
		{
			name:   "missing digest",
			status: http.StatusOK,
		},
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gets atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/" {
					return
				}
				if r.Method == http.MethodGet {
					gets.Add(1)
				}
				w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(server.Close)
			image, err := ParseImage(ParseImageOptions{Name: strings.TrimPrefix(server.URL, "https://") + "/test:latest"})
			require.NoError(t, err)
			client := newTestRegistryClient(t, Options{Timeout: 100 * time.Millisecond})
			_, _, err = client.Manifest(image, Manifest{})
			require.Error(t, err)
			assert.Zero(t, gets.Load())
		})
	}
}
