//go:build integration

package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider/docker"
	"github.com/stretchr/testify/require"
)

func TestLocalDigestDocker(t *testing.T) {
	requireLocalDigestFixture(t)
	f := newRegistryFixture(t)
	publish := runnablePublisher(t, f)
	publish("A")
	image := f.image("latest").Name
	const name = "diun-digest-running"
	runDocker(t, "run", "-d", "--name", name, "--label", "diun.enable=true", "--label", "diun.regopt=local", image)
	t.Cleanup(func() { runDocker(t, "rm", "-f", name) })
	cfg := (&model.PrdDocker{}).GetDefaults()
	cfg.Endpoint = os.Getenv("DOCKER_HOST")
	cfg.TLSVerify = new(false)
	diun := startDigestWatcher(t, f, &model.Providers{Docker: cfg})
	expectDigestMetric(t, "docker", 0, 1)
	publish("B")
	diun.Run()
	expectDigestMetric(t, "docker", 1, 2)
	runDocker(t, "pull", image)
	imageID := strings.TrimSpace(string(runCommand(t, "docker", "inspect", "--format", "{{.Image}}", name)))
	t.Logf("running image: %s", runCommand(t, "docker", "image", "inspect", "--format", "{{json .RepoDigests}} {{json .Descriptor}}", imageID))
	t.Logf("container identity: %s", runCommand(t, "docker", "inspect", "--format", "{{.Config.Image}} {{json .ImageManifestDescriptor}}", name))
	expectDigestMetric(t, "docker", 1, 2)
	diun.Run()
	expectDigestMetric(t, "docker", 1, 3)
	runDocker(t, "rm", "-f", name)
	runDocker(t, "run", "-d", "--name", name, "--label", "diun.enable=true", "--label", "diun.regopt=local", image)
	expectDigestMetric(t, "docker", 0, 3)
}

func TestDockerProvider(t *testing.T) {
	runDocker(t, "run", "-d", "--name", "watched", "--label", "diun.enable=true", "--label", "diun.watch_repo=true", "--label", "diun.include_tags=^1\\.36\\.1$", "busybox:1.36.1", "sleep", "1800")
	t.Cleanup(func() { runDocker(t, "rm", "-f", "watched") })
	for _, fixture := range []string{"docker1", "docker2", "docker3", "docker4", "docker5"} {
		t.Run(fixture, func(t *testing.T) {
			runApplicationFixture(t, fixture, "docker.io/library/busybox", "")
		})
	}
}

func TestDockerDiscovery(t *testing.T) {
	endpoint := os.Getenv("DOCKER_HOST")
	require.NotEmpty(t, endpoint, "run with test/compose.yaml against the isolated daemon")
	const image = "busybox:1.36.1"
	runDocker(t, "pull", image)
	fixtures := []struct {
		name    string
		label   string
		running bool
	}{
		{
			name:    "enabled",
			label:   "true",
			running: true,
		},
		{
			name:    "disabled",
			label:   "false",
			running: true,
		},
		{
			name:    "unlabeled",
			running: true,
		},
		{
			name:  "stopped",
			label: "true",
		},
	}
	for _, fixture := range fixtures {
		args := []string{"create", "--name", fixture.name}
		if fixture.label != "" {
			args = append(args, "--label", "diun.enable="+fixture.label)
		}
		args = append(args, image, "sleep", "600")
		runDocker(t, args...)
		t.Cleanup(func() { runDocker(t, "rm", "-f", fixture.name) })
		if fixture.running {
			runDocker(t, "start", fixture.name)
		}
	}
	cases := []struct {
		name           string
		watchByDefault bool
		watchStopped   bool
		want           []string
	}{
		{
			name: "explicit labels",
			want: []string{"enabled"},
		},
		{
			name:           "watch by default",
			watchByDefault: true,
			want:           []string{"enabled", "unlabeled"},
		},
		{
			name:         "include stopped",
			watchStopped: true,
			want:         []string{"enabled", "stopped"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&model.PrdDocker{}).GetDefaults()
			cfg.Endpoint = endpoint
			cfg.TLSVerify = new(false)
			cfg.WatchByDefault = new(tc.watchByDefault)
			cfg.WatchStopped = new(tc.watchStopped)
			jobs := docker.New(cfg, (&model.Defaults{}).GetDefaults()).ListJob()
			var names []string
			for _, job := range jobs {
				require.Equal(t, "docker", job.Provider)
				names = append(names, job.Image.Metadata["ctn_names"])
			}
			require.ElementsMatch(t, tc.want, names)
		})
	}
}

func runDocker(t *testing.T, args ...string) {
	t.Helper()
	runCommand(t, "docker", args...)
}
