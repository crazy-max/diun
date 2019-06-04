package registry

import (
	"time"

	"github.com/containers/image/manifest"
	"github.com/opencontainers/go-digest"
)

type Inspect struct {
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

// Inspect inspects a Docker image
func (c *Client) Inspect(opts *Options) (Inspect, error) {
	ctx, cancel := c.timeoutContext(opts.Timeout)
	defer cancel()

	img, _, err := c.newImage(ctx, opts)
	if err != nil {
		return Inspect{}, err
	}
	defer img.Close()

	rawManifest, _, err := img.Manifest(ctx)
	if err != nil {
		return Inspect{}, err
	}

	imgInspect, err := img.Inspect(ctx)
	if err != nil {
		return Inspect{}, err
	}

	imgDigest, err := manifest.Digest(rawManifest)
	if err != nil {
		return Inspect{}, err
	}

	imgTag := imgInspect.Tag
	if imgTag == "" {
		imgTag = opts.Image.Tag
	}

	return Inspect{
		Name:          img.Reference().DockerReference().Name(),
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
