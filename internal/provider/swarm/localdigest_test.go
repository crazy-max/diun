package swarm

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/moby/moby/api/types/swarm"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestTaskDigests(t *testing.T) {
	old, current := digest.FromString("old"), digest.FromString("current")
	cfg := (&model.PrdSwarm{}).GetDefaults()
	services := []swarm.Service{
		{
			ID: "watched",
			Spec: swarm.ServiceSpec{
				Annotations:  swarm.Annotations{Labels: map[string]string{"diun.enable": "true"}},
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest@" + current.String()}},
			},
		},
		{
			ID: "excluded",
			Spec: swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest"}},
			},
		},
	}
	tasks := []swarm.Task{
		{
			ServiceID:    "watched",
			Status:       swarm.TaskStatus{State: swarm.TaskStateRunning},
			DesiredState: swarm.TaskStateShutdown,
			Spec:         swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest@" + old.String()}},
		},
		{
			ServiceID: "watched",
			Status:    swarm.TaskStatus{State: swarm.TaskStateRunning},
			Spec:      swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest@" + current.String()}},
		},
		{
			ServiceID: "excluded",
			Status:    swarm.TaskStatus{State: swarm.TaskStateRunning},
			Spec:      swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest@" + old.String()}},
		},
		{
			ServiceID: "watched",
			Status:    swarm.TaskStatus{State: swarm.TaskStateShutdown},
			Spec:      swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: "alpine:latest@" + old.String()}},
		},
	}
	key := "docker.io/library/alpine:latest"
	require.Equal(t, [][]digest.Digest{{old}, {current}}, taskDigests(services, tasks, cfg, nil)[key])
	tasks[0].Status.State = swarm.TaskStateShutdown
	require.Equal(t, [][]digest.Digest{{current}}, taskDigests(services, tasks, cfg, nil)[key])
	tasks[1].Spec.ContainerSpec.Image = "alpine:latest"
	require.Equal(t, [][]digest.Digest{nil}, taskDigests(services, tasks, cfg, nil)[key])
	tasks[1].Status.State = swarm.TaskStateShutdown
	require.Empty(t, taskDigests(services, tasks, cfg, nil))
}
