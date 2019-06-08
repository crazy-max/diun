package docker

import (
	"fmt"
	"strings"

	"github.com/containers/image/docker"
	"github.com/containers/image/types"
)

func (c *RegistryClient) newImage(imageStr string) (types.ImageCloser, error) {
	if !strings.HasPrefix(imageStr, "//") {
		imageStr = fmt.Sprintf("//%s", imageStr)
	}

	ref, err := docker.ParseReference(imageStr)
	if err != nil {
		return nil, fmt.Errorf("invalid image name %s: %v", imageStr, err)
	}

	img, err := ref.NewImage(c.ctx, c.sysCtx)
	if err != nil {
		return nil, err
	}

	return img, nil
}
