//go:build integration

package integration

import (
	"maps"
	"os"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider/swarm"
	dockerclient "github.com/crazy-max/diun/v4/pkg/docker"
	swarmtypes "github.com/moby/moby/api/types/swarm"
	mobyclient "github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

func TestLocalDigestSwarm(t *testing.T) {
	requireLocalDigestFixture(t)
	f := newRegistryFixture(t)
	publish := runnablePublisher(t, f)
	old := publish("A")
	image := f.image("latest").Name
	runDocker(t, "swarm", "init", "--advertise-addr", "eth0")
	t.Cleanup(func() { runDocker(t, "swarm", "leave", "--force") })
	const service = "diun-digest-service"
	runDocker(t, "service", "create", "--detach=true", "--no-resolve-image", "--name", service, "--replicas", "2", "--label", "diun.enable=true", "--label", "diun.regopt=local", image+"@"+old.index.String())
	t.Cleanup(func() { runDocker(t, "service", "rm", service) })
	cfg := (&model.PrdSwarm{}).GetDefaults()
	cfg.Endpoint = os.Getenv("DOCKER_HOST")
	cfg.TLSVerify = new(false)
	cli, err := dockerclient.NewWithContext(t.Context(), dockerclient.Options{Endpoint: cfg.Endpoint})
	require.NoError(t, err)
	defer cli.Close()
	waitImages := func(want map[string]int) {
		t.Helper()
		require.Eventually(t, func() bool {
			tasks, err := cli.API.TaskList(t.Context(), mobyclient.TaskListOptions{})
			if err != nil {
				return false
			}
			got := map[string]int{}
			for _, task := range tasks.Items {
				if task.Status.State == swarmtypes.TaskStateRunning && task.Spec.ContainerSpec != nil {
					got[task.Spec.ContainerSpec.Image]++
				}
			}
			return maps.Equal(got, want)
		}, 3*time.Minute, 200*time.Millisecond, "Swarm did not reach expected running versions: %v", want)
	}
	oldRef := image + "@" + old.index.String()
	waitImages(map[string]int{oldRef: 2})
	diun := startDigestWatcher(t, f, &model.Providers{Swarm: cfg})
	expectDigestMetric(t, "swarm", 0, 1)
	current := publish("B")
	newRef := image + "@" + current.index.String()
	diun.Run()
	expectDigestMetric(t, "swarm", 1, 2)
	runDocker(t, "service", "update", "--detach=true", "--no-resolve-image", "--update-parallelism", "1", "--update-delay", "5m", "--image", newRef, service)
	waitImages(map[string]int{oldRef: 1, newRef: 1})
	expectDigestMetric(t, "swarm", 1, 2)
	runDocker(t, "service", "update", "--detach=true", "--update-delay", "0s", service)
	waitImages(map[string]int{newRef: 2})
	expectDigestMetric(t, "swarm", 0, 2)
}

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
