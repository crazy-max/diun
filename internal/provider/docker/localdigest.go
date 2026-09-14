package docker

import (
	"github.com/crazy-max/diun/v4/pkg/docker"
	"github.com/opencontainers/go-digest"
)

// LocalDigests resolves the current local image digest for each of the
// given container IDs directly from the Docker daemon. It makes no
// registry calls, so it's safe to call far more often than the watch
// schedule allows - the Prometheus recorder uses it to recompute
// diun_image_local_outdated on every scrape instead of waiting for the
// next watch run.
//
// Container IDs that can't be resolved (container removed since the last
// watch run, locally built image, dangling image, transient daemon error,
// ...) are simply omitted from the result rather than reported as a
// zero-value digest, so callers can tell "checked, matches" apart from
// "couldn't check".
func (c *Client) LocalDigests(containerIDs []string) map[string]digest.Digest {
	result := make(map[string]digest.Digest, len(containerIDs))
	if c.config == nil || len(containerIDs) == 0 {
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

	for _, id := range containerIDs {
		dgst, err := cli.ContainerImageDigest(id)
		if err != nil {
			c.logger.Debug().Err(err).Str("ctn_id", id).Msg("Cannot resolve local image digest")
			continue
		}
		result[id] = dgst
	}

	return result
}
