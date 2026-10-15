package docker

import (
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
)

// LocalDigests reports, for every container this provider watches, the
// registry digest of the image that container is running right now, keyed
// by the same normalized image reference the watch job uses (and therefore
// the same string the metrics carry in their "image" label).
//
// It re-enumerates containers on every call rather than reusing the
// container IDs captured during the last watch run. That is the whole
// point: applying an update with `docker compose up -d --pull always`
// replaces the container, so the IDs from the last watch run stop
// resolving at exactly the moment this needs to be right, and the metric
// would silently drop out instead of flipping to 0.
//
// A reference maps to every distinct digest currently in use for it, since
// several containers can share one image reference and be recreated at
// different times - the update isn't done until all of them have moved.
// References whose digest can't be resolved (locally built or dangling
// images, transient daemon errors) are left out entirely rather than
// reported as a zero value, so callers can tell "checked, matches" apart
// from "couldn't check".
//
// No registry call is made, so this is safe to run on every Prometheus
// scrape, independent of watch.schedule or registry rate limits.
func (c *Client) LocalDigests() map[string][]digest.Digest {
	result := make(map[string][]digest.Digest)
	if c.config == nil {
		return result
	}

	cli, err := docker.New(docker.Options{
		Endpoint:    c.config.Endpoint,
		APIVersion:  c.config.APIVersion,
		TLSCertPath: c.config.TLSCertsPath,
		TLSVerify:   *c.config.TLSVerify,
	})
	if err != nil {
		c.logger.Debug().Err(err).Msg("Cannot create Docker client for local digest check")
		return result
	}
	defer cli.Close()

	ctns, err := cli.ContainerList(c.containerFilter())
	if err != nil {
		c.logger.Debug().Err(err).Msg("Cannot list Docker containers for local digest check")
		return result
	}

	for _, ctn := range ctns {
		// Inspect the image the container was actually started from, not
		// the reference it was started with: a `docker pull` that hasn't
		// been followed by a recreate moves the tag while leaving the
		// container on the old image, and telling those two apart is the
		// reason this metric exists.
		imageInfo, err := cli.ImageInspect(ctn.ImageID)
		if err != nil {
			c.logger.Debug().Err(err).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", ctn.Image).
				Msg("Cannot inspect local image for digest check")
			continue
		}
		if cli.IsLocalImage(imageInfo) || cli.IsDanglingImage(imageInfo) {
			continue
		}

		dgst, err := cli.ImageDigest(imageInfo)
		if err != nil {
			c.logger.Debug().Err(err).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", ctn.Image).
				Msg("Cannot resolve local image digest")
			continue
		}

		// Mirror listContainerImage's naming so the key lines up with the
		// image the watch job recorded for this same container.
		imageName := ctn.Image
		if cli.IsDigest(imageName) {
			if len(imageInfo.RepoDigests) == 0 {
				continue
			}
			imageName = imageInfo.RepoDigests[0]
		}

		ref, err := normalizeImageRef(imageName)
		if err != nil {
			c.logger.Debug().Err(err).
				Str("ctn_id", ctn.ID).
				Str("ctn_image", imageName).
				Msg("Cannot normalize image reference for digest check")
			continue
		}

		result[ref] = appendDigest(result[ref], dgst)
	}

	return result
}

// normalizeImageRef renders an image reference the way registry.ParseImage
// does, so keys here match the image names Diun records for its metrics.
func normalizeImageRef(name string) (string, error) {
	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return "", err
	}
	return reference.TagNameOnly(named).String(), nil
}

func appendDigest(digests []digest.Digest, dgst digest.Digest) []digest.Digest {
	for _, known := range digests {
		if known == dgst {
			return digests
		}
	}
	return append(digests, dgst)
}
