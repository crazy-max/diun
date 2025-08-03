package docker

import (
	"regexp"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// ContainerInspect returns the container information.
func (c *Client) ContainerInspect(containerID string) (container.InspectResponse, error) {
	return c.API.ContainerInspect(c.ctx, containerID)
}

// IsDigest determines image it looks like a digest based image reference.
func (c *Client) IsDigest(imageID string) bool {
	return regexp.MustCompile(`^(@|sha256:|@sha256:)([0-9a-f]{64})$`).MatchString(imageID)
}

// ImageInspect returns the image information.
func (c *Client) ImageInspect(imageID string) (image.InspectResponse, error) {
	return c.API.ImageInspect(c.ctx, imageID)
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
