package docker

import (
	"regexp"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	mobyclient "github.com/moby/moby/client"
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
