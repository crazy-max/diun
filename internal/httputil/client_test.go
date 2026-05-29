package httputil

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTransport(t *testing.T) {
	t.Run("uses default proxy configuration", func(t *testing.T) {
		transport, err := NewTransport("", false, nil)
		require.NoError(t, err)
		require.NotNil(t, transport.Proxy)
		require.NotNil(t, transport.TLSClientConfig)
		require.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("with explicit proxy", func(t *testing.T) {
		transport, err := NewTransport("http://proxy.example.com:3128", false, nil)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
		require.NoError(t, err)

		proxyURL, err := transport.Proxy(req)
		require.NoError(t, err)
		require.Equal(t, "http://proxy.example.com:3128", proxyURL.String())
	})

	t.Run("invalid proxy URL", func(t *testing.T) {
		_, err := NewTransport(":", false, nil)
		require.Error(t, err)
	})

	t.Run("proxy URL without host", func(t *testing.T) {
		_, err := NewTransport("proxy.example.com:3128", false, nil)
		require.Error(t, err)
	})
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("http://proxy.example.com:3128", true, nil)
	require.NoError(t, err)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}
