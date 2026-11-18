package docker

import (
	"context"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/docker"
	mobyclient "github.com/moby/moby/client"
	"github.com/opencontainers/go-digest"
	"go.podman.io/image/v5/docker/reference"
)

// LocalDigests resolves watched containers by image reference, including all
// repository digests for each container. An empty digest set means unknown.
func LocalDigests(ctx context.Context, cfg *model.PrdDocker, defaults *model.Defaults) (map[string][][]digest.Digest, error) {
	cli, err := docker.NewWithContext(ctx, docker.Options{
		Endpoint: cfg.Endpoint, APIVersion: cfg.APIVersion,
		TLSCertPath: cfg.TLSCertsPath, TLSVerify: *cfg.TLSVerify,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	filters := make(mobyclient.Filters).Add("status", "running")
	if *cfg.WatchStopped {
		filters = filters.Add("status", "created").Add("status", "exited")
	}
	containers, err := cli.ContainerList(filters)
	if err != nil {
		return nil, err
	}
	result := make(map[string][][]digest.Digest)
	for _, ctn := range containers {
		imageName, err := cli.ContainerImageName(ctn)
		if err != nil || cli.IsDigest(imageName) {
			continue
		}
		key, err := provider.LocalImage(imageName, ctn.Labels, *cfg.WatchByDefault, defaults)
		if err != nil || key == "" {
			continue
		}
		named, err := reference.ParseNormalizedNamed(imageName)
		if err != nil {
			continue
		}
		var digests []digest.Digest
		info, err := cli.ImageInspect(ctn.ImageID)
		if err == nil {
			for _, rd := range info.RepoDigests {
				ref, err := reference.ParseNormalizedNamed(rd)
				if err != nil || ref.Name() != named.Name() {
					continue
				}
				if canonical, ok := ref.(reference.Canonical); ok {
					digests = append(digests, canonical.Digest())
				}
			}
			// Docker's containerd image store can drop repository names from
			// an old image after a pull, but retains its OCI target digest.
			if len(digests) == 0 && info.Descriptor != nil && info.Descriptor.Digest != "" {
				digests = append(digests, info.Descriptor.Digest)
			}
		}
		result[key] = append(result[key], digests)
	}
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
