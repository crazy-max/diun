package docker

import (
	"time"

	"github.com/containers/image/manifest"
	"github.com/crazy-max/diun/pkg/docker/registry"
	"github.com/opencontainers/go-digest"
)

type Manifest struct {
	Name          string
	Tag           string
	MIMEType      string
	Digest        digest.Digest
	Created       *time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
}

// Manifest returns the manifest for a specific image
func (c *RegistryClient) Manifest(image registry.Image) (Manifest, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	imgCls, err := c.newImage(ctx, image.String())
	if err != nil {
		return Manifest{}, err
	}
	defer imgCls.Close()

	rawManifest, _, err := imgCls.Manifest(ctx)
	if err != nil {
		return Manifest{}, err
	}

	imgInspect, err := imgCls.Inspect(ctx)
	if err != nil {
		return Manifest{}, err
	}

	imgDigest, err := manifest.Digest(rawManifest)
	if err != nil {
		return Manifest{}, err
	}

	imgTag := imgInspect.Tag
	if imgTag == "" {
		imgTag = image.Tag
	}

	return Manifest{
		Name:          imgCls.Reference().DockerReference().Name(),
		Tag:           imgTag,
		MIMEType:      manifest.GuessMIMEType(rawManifest),
		Digest:        imgDigest,
		Created:       imgInspect.Created,
		DockerVersion: imgInspect.DockerVersion,
		Labels:        imgInspect.Labels,
		Architecture:  imgInspect.Architecture,
		Os:            imgInspect.Os,
		Layers:        imgInspect.Layers,
	}, nil
}
