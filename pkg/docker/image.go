package docker

// IsLocalImage checks if the image has been built locally
func (c *Client) IsLocalImage(image string) (bool, error) {
	raw, _, err := c.API.ImageInspectWithRaw(c.ctx, image)
	if err != nil {
		return false, err
	}
	return len(raw.RepoDigests) == 0, nil
}
