package swarm

import (
	"context"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/distribution/reference"
	"github.com/moby/moby/api/types/swarm"
	mobyclient "github.com/moby/moby/client"
	"github.com/opencontainers/go-digest"
)

// LocalDigests compares running tasks, including old tasks in a rolling update.
// The service template selects the workload; it does not prove deployment.
func LocalDigests(ctx context.Context, cfg *model.PrdSwarm, defaults *model.Defaults) (map[string][][]digest.Digest, error) {
	cli, err := docker.NewWithContext(ctx, docker.Options{
		Endpoint: cfg.Endpoint, APIVersion: cfg.APIVersion,
		TLSCertPath: cfg.TLSCertsPath, TLSVerify: *cfg.TLSVerify,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	services, err := cli.ServiceList(make(mobyclient.Filters))
	if err != nil {
		return nil, err
	}
	tasks, err := cli.API.TaskList(ctx, mobyclient.TaskListOptions{})
	if err != nil {
		return nil, err
	}
	return taskDigests(services, tasks.Items, cfg, defaults), nil
}

func taskDigests(services []swarm.Service, tasks []swarm.Task, cfg *model.PrdSwarm, defaults *model.Defaults) map[string][][]digest.Digest {
	selected := make(map[string]string)
	for _, svc := range services {
		if svc.Spec.TaskTemplate.ContainerSpec == nil {
			continue
		}
		key, err := provider.LocalImage(svc.Spec.TaskTemplate.ContainerSpec.Image, svc.Spec.Labels, *cfg.WatchByDefault, defaults)
		if err == nil && key != "" {
			selected[svc.ID] = key
		}
	}
	result := make(map[string][][]digest.Digest)
	for _, task := range tasks {
		key, ok := selected[task.ServiceID]
		if !ok || task.Status.State != swarm.TaskStateRunning {
			continue
		}
		var digests []digest.Digest
		if task.Spec.ContainerSpec != nil {
			ref, err := reference.ParseNormalizedNamed(task.Spec.ContainerSpec.Image)
			if err == nil {
				if canonical, ok := ref.(reference.Canonical); ok {
					digests = append(digests, canonical.Digest())
				}
			}
		}
		result[key] = append(result[key], digests)
	}
	return result
}
