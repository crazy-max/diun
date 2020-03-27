package registry

import (
	"fmt"
	"time"

	"github.com/containers/image/v5/manifest"
	"github.com/crazy-max/diun/pkg/registry/platform"
	"github.com/opencontainers/go-digest"
)

// Manifest is the Docker image manifest information
type Manifest struct {
	Name          string
	Tag           string
	MIMEType      string
	Digest        digest.Digest
	Created       *time.Time
	DockerVersion string
	Labels        map[string]string
	Layers        []string
	Platform      string `json:"-"`
}

// Manifest returns the manifest for a specific image
func (c *Client) Manifest(image Image) (Manifest, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgCloser, err := c.newImage(ctx, image.String())
	if err != nil {
		return Manifest{}, err
	}
	defer imgCloser.Close()

	rawManifest, _, err := imgCloser.Manifest(ctx)
	if err != nil {
		return Manifest{}, err
	}

	imgInspect, err := imgCloser.Inspect(ctx)
	if err != nil {
		return Manifest{}, err
	}

	imgDigest, err := manifest.Digest(rawManifest)
	if err != nil {
		return Manifest{}, err
	}

	platforms, err := platform.WantedPlatforms(c.sysCtx)
	if err != nil {
		return Manifest{}, err
	}

	imgTag := imgInspect.Tag
	if imgTag == "" {
		imgTag = image.Tag
	}

	imgPlatform := fmt.Sprintf("%s/%s", platforms[0].OS, platforms[0].Architecture)
	if platforms[0].Variant != "" {
		imgPlatform = fmt.Sprintf("%s/%s", imgPlatform, platforms[0].Variant)
	}

	return Manifest{
		Name:          imgCloser.Reference().DockerReference().Name(),
		Tag:           imgTag,
		MIMEType:      manifest.GuessMIMEType(rawManifest),
		Digest:        imgDigest,
		Created:       imgInspect.Created,
		DockerVersion: imgInspect.DockerVersion,
		Labels:        imgInspect.Labels,
		Layers:        imgInspect.Layers,
		Platform:      imgPlatform,
	}, nil
}
