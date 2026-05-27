package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerBearerToken(t *testing.T) {
	_, registry := NewRecorder("test")
	server, err := NewServer(&model.Metrics{
		Addr:  "127.0.0.1:0",
		Path:  "/metrics",
		Token: "secret",
	}, registry)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	server.httpServer.Handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, `Bearer realm="diun metrics"`, res.Header().Get("WWW-Authenticate"))

	res = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	server.httpServer.Handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `diun_build_info{version="test"} 1`)
}

func TestServerTokenFile(t *testing.T) {
	tokenFile := t.TempDir() + "/metrics-token"
	err := os.WriteFile(tokenFile, []byte("secret"), 0o600)
	require.NoError(t, err)

	_, registry := NewRecorder("test")
	server, err := NewServer(&model.Metrics{
		Addr:      "127.0.0.1:0",
		Path:      "/metrics",
		TokenFile: tokenFile,
	}, registry)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	server.httpServer.Handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "diun_build_info")
}
