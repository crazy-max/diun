package registry

import (
	"fmt"
	"time"

	"github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"go.podman.io/image/v5/docker"
	"go.podman.io/image/v5/manifest"
)

// Manifest is the Docker image manifest information
type Manifest struct {
	Name           string
	Tag            string
	MIMEType       string
	Digest         digest.Digest
	InstanceDigest digest.Digest `json:",omitempty"`
	Created        *time.Time
	DockerVersion  string
	Labels         map[string]string
	Layers         []string
	Platform       string
	Raw            []byte
}

// Manifest returns the manifest for a specific image
func (c *Client) Manifest(image Image, dbManifest Manifest) (Manifest, bool, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	rmRef, err := ImageReference(image.String())
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot parse reference")
	}

	// Retrieve remote digest through HEAD request
	rmDigest, err := docker.GetDigest(ctx, c.sysCtx, rmRef)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image digest from HEAD request")
	}

	// Digest match, returns db manifest
	if c.opts.CompareDigest && len(dbManifest.Digest) > 0 && dbManifest.Digest == rmDigest {
		if !dbManifest.isManifestList() {
			return dbManifest, false, nil
		}
		// The index digest alone cannot identify the cached platform's metadata.
		// Older records lack InstanceDigest and are refreshed once.
		list, err := manifest.ListFromBlob(dbManifest.Raw, dbManifest.MIMEType)
		if err == nil {
			selected, err := list.ChooseInstance(c.sysCtx)
			if err == nil && selected == dbManifest.InstanceDigest {
				instance, err := list.Instance(selected)
				// A nested index still needs platform selection within the child.
				if err == nil && !isManifestList(instance.MediaType) {
					return dbManifest, false, nil
				}
			}
		}
	}

	rmCloser, err := rmRef.NewImage(ctx, c.sysCtx)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot create image closer")
	}
	defer rmCloser.Close()

	rmRawManifest, rmManifestMimeType, err := rmCloser.Manifest(ctx)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get raw manifest")
	}

	var instanceDigest digest.Digest
	if isManifestList(rmManifestMimeType) {
		list, err := manifest.ListFromBlob(rmRawManifest, rmManifestMimeType)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot parse manifest list")
		}
		instanceDigest, err = list.ChooseInstance(c.sysCtx)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "error choosing image instance")
		}
	}

	// For manifests list compare also digest matching the platform
	updated := dbManifest.Digest != rmDigest
	if c.opts.CompareDigest && len(dbManifest.Raw) > 0 && dbManifest.isManifestList() && isManifestList(rmManifestMimeType) {
		dbManifestList, err := manifest.ListFromBlob(dbManifest.Raw, dbManifest.MIMEType)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot parse manifest list")
		}
		dbManifestPlatformDigest, err := dbManifestList.ChooseInstance(c.sysCtx)
		// A newly published platform has no instance in the previous index.
		// The current index was successfully resolved above, so it is an update.
		updated = err != nil || dbManifestPlatformDigest != instanceDigest
	}

	// Metadata describing the Docker image
	rmInspect, err := rmCloser.Inspect(ctx)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot inspect")
	}
	rmTag := rmInspect.Tag
	if len(rmTag) == 0 {
		rmTag = image.Tag
	}
	rmPlatform := fmt.Sprintf("%s/%s", rmInspect.Os, rmInspect.Architecture)
	if rmInspect.Variant != "" {
		rmPlatform = fmt.Sprintf("%s/%s", rmPlatform, rmInspect.Variant)
	}

	return Manifest{
		Name:           rmCloser.Reference().DockerReference().Name(),
		Tag:            rmTag,
		MIMEType:       rmManifestMimeType,
		Digest:         rmDigest,
		InstanceDigest: instanceDigest,
		Created:        rmInspect.Created,
		DockerVersion:  rmInspect.DockerVersion,
		Labels:         rmInspect.Labels,
		Layers:         rmInspect.Layers,
		Platform:       rmPlatform,
		Raw:            rmRawManifest,
	}, updated, nil
}

func (m Manifest) isManifestList() bool {
	return isManifestList(m.MIMEType)
}

func isManifestList(mimeType string) bool {
	return mimeType == manifest.DockerV2ListMediaType || mimeType == ocispecs.MediaTypeImageIndex
}
