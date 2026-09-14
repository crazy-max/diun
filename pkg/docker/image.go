package docker

import (
	"regexp"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	mobyclient "github.com/moby/moby/client"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

// ContainerInspect returns the container information.
func (c *Client) ContainerInspect(containerID string) (container.InspectResponse, error) {
	result, err := c.API.ContainerInspect(c.ctx, containerID, mobyclient.ContainerInspectOptions{})
	if err != nil {
		return container.InspectResponse{}, err
	}
	return result.Container, nil
}

// IsDigest determines image it looks like a digest based image reference.
func (c *Client) IsDigest(imageID string) bool {
	return regexp.MustCompile(`^(@|sha256:|@sha256:)([0-9a-f]{64})$`).MatchString(imageID)
}

// ImageInspect returns the image information.
func (c *Client) ImageInspect(imageID string) (image.InspectResponse, error) {
	result, err := c.API.ImageInspect(c.ctx, imageID)
	if err != nil {
		return image.InspectResponse{}, err
	}
	return result.InspectResponse, nil
}

// IsLocalImage checks if the image has been built locally
func (c *Client) IsLocalImage(image image.InspectResponse) bool {
	return len(image.RepoDigests) == 0
}

// IsDanglingImage returns whether the given image is "dangling" which means
// that there are no repository references to the given image and it has no
// child images
func (c *Client) IsDanglingImage(image image.InspectResponse) bool {
	return len(image.RepoTags) == 1 && image.RepoTags[0] == "<none>:<none>" && len(image.RepoDigests) == 1 && image.RepoDigests[0] == "<none>@<none>"
}

// ImageDigest extracts the content digest from an image's repo digests
// (e.g. "alpine@sha256:...") as reported by the local image store.
func (c *Client) ImageDigest(img image.InspectResponse) (digest.Digest, error) {
	for _, repoDigest := range img.RepoDigests {
		if _, dgst, ok := strings.Cut(repoDigest, "@"); ok && len(dgst) > 0 {
			return digest.Digest(dgst), nil
		}
	}
	return "", errors.Errorf("no repo digest found in %v", img.RepoDigests)
}

// ContainerImageDigest returns the digest of the image currently backing
// the given container, resolved entirely from the local Docker daemon
// (container inspect + image inspect). It makes no registry calls.
func (c *Client) ContainerImageDigest(containerID string) (digest.Digest, error) {
	ctn, err := c.ContainerInspect(containerID)
	if err != nil {
		return "", err
	}

	imageInfo, err := c.ImageInspect(ctn.Image)
	if err != nil {
		return "", err
	}

	return c.ImageDigest(imageInfo)
}
