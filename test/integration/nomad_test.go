//go:build integration

package integration

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/require"
)

func TestNomadProvider(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.HttpClient = &http.Client{Timeout: 10 * time.Second}
	client, err := api.NewClient(cfg)
	require.NoError(t, err)
	data, err := os.ReadFile("/src/test/fixtures/nomad1/job.nomad.hcl")
	require.NoError(t, err)
	job, err := client.Jobs().ParseHCL(string(data), true)
	require.NoError(t, err)
	_, _, err = client.Jobs().Register(job, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _, err := client.Jobs().Deregister(*job.ID, true, nil)
		require.NoError(t, err)
	})
	runApplicationFixture(t, "nomad1", "docker.io/library/busybox", "1.36.1")
}
