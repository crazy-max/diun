package kubernetes

import (
	"context"
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/k8s"
	"github.com/opencontainers/go-digest"
	"go.podman.io/image/v5/docker/reference"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocalDigests inspects every Pod replica, using runtime status rather than
// desired image tags. Opaque runtime IDs are not registry manifest digests.
func LocalDigests(ctx context.Context, cfg *model.PrdKubernetes, defaults *model.Defaults) (map[string][][]digest.Digest, error) {
	cli, err := k8s.NewWithContext(ctx, k8s.Options{
		Endpoint: cfg.Endpoint, Token: cfg.Token, TokenFile: cfg.TokenFile,
		CertAuthFilePath: cfg.CertAuthFilePath, TLSInsecure: cfg.TLSInsecure,
		Namespaces: cfg.Namespaces,
	})
	if err != nil {
		return nil, err
	}
	pods, err := cli.PodListAll(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podDigests(pods, cfg, defaults), nil
}

func podDigests(pods []v1.Pod, cfg *model.PrdKubernetes, defaults *model.Defaults) map[string][][]digest.Digest {
	result := make(map[string][][]digest.Digest)
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}
		for _, ctn := range pod.Spec.Containers {
			key, err := provider.LocalImage(ctn.Image, pod.Annotations, *cfg.WatchByDefault, defaults)
			if err != nil || key == "" {
				continue
			}
			var digests []digest.Digest
			for _, status := range pod.Status.ContainerStatuses {
				if status.Name != ctn.Name || status.State.Running == nil {
					continue
				}
				// docker://sha256:... and bare IDs can denote config digests.
				// Only repository-qualified, pullable references are comparable.
				id := strings.TrimPrefix(status.ImageID, "docker-pullable://")
				ref, err := reference.ParseNormalizedNamed(id)
				if err != nil {
					continue
				}
				if canonical, ok := ref.(reference.Canonical); ok {
					digests = append(digests, canonical.Digest())
				}
			}
			result[key] = append(result[key], digests)
		}
	}
	return result
}
