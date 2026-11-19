package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker/schema1"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
)

// Manifest is the Docker image manifest information.
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

// NonImageArtifactError identifies OCI artifacts that should not be monitored.
type NonImageArtifactError struct {
	MediaType string
}

func (e NonImageArtifactError) Error() string {
	return fmt.Sprintf("non-image artifact with config media type %q", e.MediaType)
}

// Manifest checks the tag digest before fetching image metadata.
func (c *Client) Manifest(image Image, dbManifest Manifest) (Manifest, bool, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	var result Manifest
	var updated bool
	err := retry(ctx, func() error {
		var err error
		result, updated, err = c.manifest(ctx, image, dbManifest)
		return err
	})

	return result, updated, err
}

func (c *Client) manifest(ctx context.Context, image Image, dbManifest Manifest) (Manifest, bool, error) {
	r, err := ImageReference(image.String())
	if err != nil {
		return Manifest{}, false, err
	}
	clients, err := c.registryClient(ctx, r.Registry)
	if err != nil {
		return Manifest{}, false, err
	}
	head, err := clients.head.ManifestHead(ctx, r)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image digest from HEAD request")
	}
	remoteDigest := head.GetDescriptor().Digest
	if remoteDigest == "" {
		return Manifest{}, false, errors.New("manifest HEAD request returned empty digest")
	}
	if c.opts.CompareDigest && dbManifest.Digest == remoteDigest {
		if !dbManifest.isManifestList() {
			return dbManifest, false, nil
		}
		stored, err := parseManifest(dbManifest.Raw, dbManifest.MIMEType)
		if err == nil {
			desc, err := manifest.GetPlatformDesc(stored, &c.platform)
			if err == nil && desc.Digest == dbManifest.InstanceDigest && !isManifestList(desc.MediaType) {
				return dbManifest, false, nil
			}
		}
	}
	// Pin to the HEAD digest: aliases share immutable cached manifests and a
	// moving tag cannot associate metadata with a different digest.
	client := clients.content
	top, err := client.ManifestGet(ctx, r.SetDigest(remoteDigest.String()))
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get manifest")
	}
	raw, err := top.RawBody()
	if err != nil {
		return Manifest{}, false, err
	}
	selected := top
	var instanceDigest digest.Digest
	seen := map[digest.Digest]bool{remoteDigest: true}
	for selected.IsList() {
		desc, err := manifest.GetPlatformDesc(selected, &c.platform)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "error choosing image instance")
		}
		if instanceDigest == "" {
			instanceDigest = desc.Digest
		}
		if seen[desc.Digest] || len(seen) >= 32 {
			return Manifest{}, false, errors.New("manifest index cycle or excessive nesting")
		}
		seen[desc.Digest] = true
		selected, err = client.ManifestGet(ctx, r, regclient.WithManifestDesc(*desc))
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot get selected platform manifest")
		}
	}
	updated := dbManifest.Digest != remoteDigest
	if c.opts.CompareDigest && len(dbManifest.Raw) > 0 && dbManifest.isManifestList() && top.IsList() {
		stored, err := parseManifest(dbManifest.Raw, dbManifest.MIMEType)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot parse stored manifest")
		}
		desc, err := manifest.GetPlatformDesc(stored, &c.platform)
		updated = err != nil || desc.Digest != instanceDigest
	}
	result := Manifest{Name: image.Name(), Tag: image.Tag, MIMEType: top.GetDescriptor().MediaType, Digest: remoteDigest, InstanceDigest: instanceDigest, Raw: raw}
	if isSchema1(selected.GetDescriptor().MediaType) {
		if err := inspectSchema1(selected, &result); err != nil {
			return Manifest{}, false, err
		}
		return result, updated, nil
	}
	im, ok := selected.(manifest.Imager)
	if !ok {
		return Manifest{}, false, NonImageArtifactError{MediaType: selected.GetDescriptor().MediaType}
	}
	configDesc, err := im.GetConfig()
	if err != nil {
		return Manifest{}, false, err
	}
	if selected.GetDescriptor().MediaType == mediatype.OCI1Manifest && configDesc.MediaType != mediatype.OCI1ImageConfig {
		return Manifest{}, false, NonImageArtifactError{MediaType: configDesc.MediaType}
	}
	// Regclient's limited reader adds one to this size, overflowing at MaxInt64.
	if configDesc.Size == math.MaxInt64 {
		return Manifest{}, false, errors.Errorf("invalid image config size: %d", configDesc.Size)
	}
	layers, err := im.GetLayers()
	if err != nil {
		return Manifest{}, false, err
	}
	config, err := client.BlobGetOCIConfig(ctx, r, configDesc)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image config")
	}
	configRaw, err := config.RawBody()
	if err != nil {
		return Manifest{}, false, err
	}
	if err := inspectConfig(configRaw, &result); err != nil {
		return Manifest{}, false, err
	}
	for _, layer := range layers {
		result.Layers = append(result.Layers, layer.Digest.String())
	}
	return result, updated, nil
}

func inspectConfig(raw []byte, result *Manifest) error {
	var config struct {
		Created       *time.Time
		DockerVersion string `json:"docker_version"`
		OS            string
		Architecture  string
		Variant       string
		Config        struct{ Labels map[string]string }
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return errors.Wrap(err, "cannot inspect image config")
	}
	result.Created, result.DockerVersion, result.Labels = config.Created, config.DockerVersion, config.Config.Labels
	result.Platform = config.OS + "/" + config.Architecture
	if config.Variant != "" {
		result.Platform += "/" + config.Variant
	}
	return nil
}

func inspectSchema1(m manifest.Manifest, result *Manifest) error {
	var s schema1.Manifest
	switch v := m.GetOrig().(type) {
	case schema1.Manifest:
		s = v
	case schema1.SignedManifest:
		s = v.Manifest
	default:
		return errors.New("invalid schema1 manifest")
	}
	if len(s.History) == 0 || len(s.History) != len(s.FSLayers) {
		return errors.New("invalid schema1 history")
	}
	if err := inspectConfig([]byte(s.History[0].V1Compatibility), result); err != nil {
		return err
	}
	if s.Tag != "" {
		result.Tag = s.Tag
	}
	for _, layer := range slices.Backward(s.FSLayers) {
		result.Layers = append(result.Layers, layer.BlobSum.String())
	}
	return nil
}

func parseManifest(raw []byte, mediaType string) (manifest.Manifest, error) {
	return manifest.New(manifest.WithRaw(raw), manifest.WithDesc(descriptor.Descriptor{MediaType: mediaType}))
}

func (m Manifest) isManifestList() bool {
	return isManifestList(m.MIMEType)
}

func isManifestList(mt string) bool {
	return mt == mediatype.Docker2ManifestList || mt == mediatype.OCI1ManifestList
}

func isSchema1(mt string) bool {
	return mt == mediatype.Docker1Manifest || mt == mediatype.Docker1ManifestSigned
}
