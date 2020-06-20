package registry

import (
	"fmt"
	"time"

	"github.com/containers/image/v5/manifest"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
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
		return Manifest{}, errors.Wrap(err, "Cannot create image closer")
	}
	defer imgCloser.Close()

	rawManifest, _, err := imgCloser.Manifest(ctx)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot get raw manifest")
	}

	imgInspect, err := imgCloser.Inspect(ctx)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot inspect")
	}

	imgDigest, err := manifest.Digest(rawManifest)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot get digest")
	}

	imgTag := imgInspect.Tag
	if len(imgTag) == 0 {
		imgTag = image.Tag
	}

	imgPlatform := fmt.Sprintf("%s/%s", imgInspect.Os, imgInspect.Architecture)
	if imgInspect.Variant != "" {
		imgPlatform = fmt.Sprintf("%s/%s", imgPlatform, imgInspect.Variant)
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
