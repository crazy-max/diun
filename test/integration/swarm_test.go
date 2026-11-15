//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider/swarm"
	"github.com/stretchr/testify/require"
)

func TestSwarmProvider(t *testing.T) {
	endpoint := os.Getenv("DOCKER_HOST")
	require.NotEmpty(t, endpoint, "run with test/compose.yaml against the isolated daemon")
	runDocker(t, "swarm", "init", "--advertise-addr", "eth0")
	t.Cleanup(func() { runDocker(t, "swarm", "leave", "--force") })
	// Discovery reads service specs, including services scaled to zero.
	runDocker(t, "stack", "deploy", "--detach=true", "--resolve-image", "never", "--compose-file", "/src/test/fixtures/swarm1/stack.yml", "diun-integration")
	t.Cleanup(func() { runDocker(t, "stack", "rm", "diun-integration") })

	cases := []struct {
		name           string
		watchByDefault bool
		want           []string
	}{
		{
			name: "service labels",
			want: []string{"diun-integration_enabled"},
		},
		{
			name:           "watch by default",
			watchByDefault: true,
			want:           []string{"diun-integration_enabled", "diun-integration_unlabeled", "diun-integration_container-label"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&model.PrdSwarm{}).GetDefaults()
			cfg.Endpoint = endpoint
			cfg.TLSVerify = new(false)
			cfg.WatchByDefault = new(tc.watchByDefault)
			jobs := swarm.New(cfg, (&model.Defaults{}).GetDefaults()).ListJob()
			var names []string
			for _, job := range jobs {
				require.Equal(t, "swarm", job.Provider)
				require.Equal(t, "busybox:1.36.1", job.Image.Name)
				require.NotEmpty(t, job.Image.Metadata["svc_id"])
				names = append(names, job.Image.Metadata["ctn_name"])
			}
			require.ElementsMatch(t, tc.want, names)
		})
	}
	t.Run("application", func(t *testing.T) {
		runApplicationFixture(t, "swarm1", "docker.io/library/busybox", "1.36.1")
	})
}
