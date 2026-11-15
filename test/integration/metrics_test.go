//go:build integration

package integration

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	runApplicationFixture(t, "metrics", "docker.io/library/alpine", "")
	checkMetrics(t)
}

func checkMetrics(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	for _, authorized := range []bool{false, true} {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:9090/metrics", nil)
		require.NoError(t, err)
		if authorized {
			req.Header.Set("Authorization", "Bearer diun-metrics-e2e-token")
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		if !authorized {
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			continue
		}
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, string(body), "diun_build_info{version=")
		require.Contains(t, string(body), "diun_watch_runs_total 1\n")
		require.Contains(t, string(body), "diun_watch_last_run_timestamp_seconds ")
	}
}
