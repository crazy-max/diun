package registry

import (
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/regclient/regclient"
	regdescriptor "github.com/regclient/regclient/types/descriptor"
	regmanifest "github.com/regclient/regclient/types/manifest"
	regplatform "github.com/regclient/regclient/types/platform"
)

type Manifest struct {
	Name     string
	Tag      string
	MIMEType string
	Digest   digest.Digest
	Created  *time.Time
	Labels   map[string]string
	Layers   []string
	Platform string
	Raw      []byte
}

func (c *Client) Manifest(image Image, dbManifest Manifest) (Manifest, bool, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()

	regRef, err := image.regRef()
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot create regclient reference")
	}

	headManifest, err := c.regctl.ManifestHead(ctx, regRef)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get manifest digest from HEAD request")
	} else if headManifest == nil || headManifest.GetDescriptor().Digest == "" {
		return Manifest{}, false, errors.New("manifest HEAD request returned no manifest or empty digest")
	}

	remoteDigest := headManifest.GetDescriptor().Digest
	if c.opts.CompareDigest && len(dbManifest.Digest) > 0 && dbManifest.Digest == remoteDigest {
		return dbManifest, false, nil
	}

	topManifest, err := c.regctl.ManifestGet(ctx, regRef)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get manifest")
	}
	remoteDigest = topManifest.GetDescriptor().Digest

	remoteRawManifest, err := topManifest.RawBody()
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot read raw manifest")
	}
	remoteManifestMediaType := topManifest.GetDescriptor().MediaType

	updated := dbManifest.Digest != remoteDigest
	if c.opts.CompareDigest && len(dbManifest.Raw) > 0 && topManifest.IsList() {
		dbManifestValue, err := parseManifest(dbManifest.Raw, dbManifest.MIMEType)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot parse stored manifest")
		}
		if dbManifestValue.IsList() {
			dbManifestPlatformDigest, err := manifestPlatformDigest(dbManifestValue, c.opts.Platform)
			if err != nil {
				return Manifest{}, false, errors.Wrap(err, "cannot choose platform digest from stored manifest list")
			}
			remoteManifestPlatformDigest, err := manifestPlatformDigest(topManifest, c.opts.Platform)
			if err != nil {
				return Manifest{}, false, errors.Wrap(err, "cannot choose platform digest from remote manifest list")
			}
			updated = dbManifestPlatformDigest != remoteManifestPlatformDigest
		}
	}

	selectedManifest := topManifest
	platform := c.opts.Platform

	if topManifest.IsList() {
		desc, err := manifestPlatformDescriptor(topManifest, platform)
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "error choosing image instance")
		}
		selectedManifest, err = c.regctl.ManifestGet(ctx, regRef, regclient.WithManifestDesc(desc))
		if err != nil {
			return Manifest{}, false, errors.Wrap(err, "cannot get selected platform manifest")
		}
		if desc.Platform != nil {
			platform = *desc.Platform
		}
	}

	imageManifest, ok := selectedManifest.(regmanifest.Imager)
	if !ok {
		return Manifest{}, false, errors.Errorf("manifest media type %q is not an image manifest", selectedManifest.GetDescriptor().MediaType)
	}

	layersDesc, err := imageManifest.GetLayers()
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image layers")
	}

	configDesc, err := imageManifest.GetConfig()
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image config descriptor")
	}

	configData, err := c.regctl.BlobGetOCIConfig(ctx, regRef, configDesc)
	if err != nil {
		return Manifest{}, false, errors.Wrap(err, "cannot get image config")
	}

	layers := make([]string, 0, len(layersDesc))
	for _, layer := range layersDesc {
		layers = append(layers, layer.Digest.String())
	}

	imageConfig := configData.GetConfig()

	remotePlatform := platform.String()
	if imageConfig.OS != "" && imageConfig.Architecture != "" {
		remotePlatform = regplatform.Platform{
			OS:           imageConfig.OS,
			Architecture: imageConfig.Architecture,
			Variant:      imageConfig.Variant,
		}.String()
	}

	return Manifest{
		Name:     image.Name(),
		Tag:      image.Tag,
		MIMEType: remoteManifestMediaType,
		Digest:   remoteDigest,
		Created:  imageConfig.Created,
		Labels:   imageConfig.Config.Labels,
		Layers:   layers,
		Platform: remotePlatform,
		Raw:      remoteRawManifest,
	}, updated, nil
}

func platformDigestFromManifest(raw []byte, mimeType string, platform regplatform.Platform) (digest.Digest, error) {
	manifest, err := parseManifest(raw, mimeType)
	if err != nil {
		return "", err
	}
	return manifestPlatformDigest(manifest, platform)
}

func manifestPlatformDigest(manifest regmanifest.Manifest, platform regplatform.Platform) (digest.Digest, error) {
	desc, err := manifestPlatformDescriptor(manifest, platform)
	if err != nil {
		return "", err
	}
	return desc.Digest, nil
}

func manifestPlatformDescriptor(manifest regmanifest.Manifest, platform regplatform.Platform) (regdescriptor.Descriptor, error) {
	desc, err := regmanifest.GetPlatformDesc(manifest, &platform)
	if err != nil {
		return regdescriptor.Descriptor{}, errors.Wrap(err, "cannot select platform descriptor")
	}
	return *desc, nil
}

func parseManifest(raw []byte, mimeType string) (regmanifest.Manifest, error) {
	opts := []regmanifest.Opts{regmanifest.WithRaw(raw)}
	if mimeType != "" {
		opts = append(opts, regmanifest.WithDesc(regdescriptor.Descriptor{MediaType: mimeType}))
	}
	return regmanifest.New(opts...)
}
