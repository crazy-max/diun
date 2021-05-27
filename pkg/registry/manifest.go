package registry

import (
	"fmt"
	"time"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
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
	Platform      string
	Raw           []byte
}

// Manifest returns the manifest for a specific image
func (c *Client) Manifest(image Image, dbManifest Manifest) (Manifest, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	if c.sysCtx.DockerAuthConfig == nil {
		c.sysCtx.DockerAuthConfig = &types.DockerAuthConfig{}
		// TODO: Seek credentials
		//auth, err := config.GetCredentials(c.sysCtx, reference.Domain(ref.DockerReference()))
		//if err != nil {
		//	return nil, errors.Wrap(err, "Cannot get registry credentials")
		//}
		//*c.sysCtx.DockerAuthConfig = auth
	}

	imgRef, err := ParseReference(image.String())
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot parse reference")
	}

	var imgDigest digest.Digest
	if c.opts.CompareDigest {
		imgDigest, err = docker.GetDigest(ctx, c.sysCtx, imgRef)
		if err != nil {
			return Manifest{}, errors.Wrap(err, "Cannot get image digest from HEAD request")
		}

		if dbManifest.Digest != "" && dbManifest.Digest == imgDigest {
			return dbManifest, nil
		}
	}

	imgCloser, err := imgRef.NewImage(ctx, c.sysCtx)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot create image closer")
	}
	defer imgCloser.Close()

	rawManifest, _, err := imgCloser.Manifest(ctx)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot get raw manifest")
	}

	if !c.opts.CompareDigest {
		imgDigest, err = manifest.Digest(rawManifest)
		if err != nil {
			return Manifest{}, errors.Wrap(err, "Cannot get digest")
		}
	}

	imgInspect, err := imgCloser.Inspect(ctx)
	if err != nil {
		return Manifest{}, errors.Wrap(err, "Cannot inspect")
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
		Raw:           rawManifest,
	}, nil
}
