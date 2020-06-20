package docker

import "github.com/docker/docker/api/types"

// RawImage returns the image information and its raw representation
func (c *Client) RawImage(image string) (types.ImageInspect, error) {
	imageRaw, _, err := c.API.ImageInspectWithRaw(c.ctx, image)
	return imageRaw, err
}

// IsLocalImage checks if the image has been built locally
func (c *Client) IsLocalImage(image types.ImageInspect) bool {
	return len(image.RepoDigests) == 0
}

// IsDanglingImage returns whether the given image is "dangling" which means
// that there are no repository references to the given image and it has no
// child images
func (c *Client) IsDanglingImage(image types.ImageInspect) bool {
	return len(image.RepoTags) == 1 && image.RepoTags[0] == "<none>:<none>" && len(image.RepoDigests) == 1 && image.RepoDigests[0] == "<none>@<none>"
}
