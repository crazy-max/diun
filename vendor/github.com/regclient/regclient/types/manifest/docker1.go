package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	digest "github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker/schema1"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/platform"
)

const (
	// MediaTypeDocker1Manifest deprecated media type for docker schema1 manifests.
	MediaTypeDocker1Manifest = "application/vnd.docker.distribution.manifest.v1+json"
	// MediaTypeDocker1ManifestSigned is a deprecated schema1 manifest with jws signing.
	MediaTypeDocker1ManifestSigned = "application/vnd.docker.distribution.manifest.v1+prettyjws"
)

type docker1Manifest struct {
	common
	schema1.Manifest
}
type docker1SignedManifest struct {
	common
	schema1.SignedManifest
}

func (m *docker1Manifest) GetConfig() (descriptor.Descriptor, error) {
	return descriptor.Descriptor{}, fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) GetConfigDigest() (digest.Digest, error) {
	return "", fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetConfig() (descriptor.Descriptor, error) {
	return descriptor.Descriptor{}, fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetConfigDigest() (digest.Digest, error) {
	return "", fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) GetManifestList() ([]descriptor.Descriptor, error) {
	return []descriptor.Descriptor{}, fmt.Errorf("platform descriptor list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetManifestList() ([]descriptor.Descriptor, error) {
	return []descriptor.Descriptor{}, fmt.Errorf("platform descriptor list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) GetLayers() ([]descriptor.Descriptor, error) {
	if !m.manifSet {
		return []descriptor.Descriptor{}, errs.ErrManifestNotSet
	}

	var dl []descriptor.Descriptor
	for _, sd := range m.FSLayers {
		dl = append(dl, descriptor.Descriptor{
			Digest: sd.BlobSum,
		})
	}
	return dl, nil
}

func (m *docker1SignedManifest) GetLayers() ([]descriptor.Descriptor, error) {
	if !m.manifSet {
		return []descriptor.Descriptor{}, errs.ErrManifestNotSet
	}

	var dl []descriptor.Descriptor
	for _, sd := range m.FSLayers {
		dl = append(dl, descriptor.Descriptor{
			Digest: sd.BlobSum,
		})
	}
	return dl, nil
}

func (m *docker1Manifest) GetOrig() any {
	return m.Manifest
}

func (m *docker1SignedManifest) GetOrig() any {
	return m.SignedManifest
}

func (m *docker1Manifest) GetPlatformDesc(p *platform.Platform) (*descriptor.Descriptor, error) {
	return nil, fmt.Errorf("platform lookup not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetPlatformDesc(p *platform.Platform) (*descriptor.Descriptor, error) {
	return nil, fmt.Errorf("platform lookup not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) GetPlatformList() ([]*platform.Platform, error) {
	return nil, fmt.Errorf("platform list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetPlatformList() ([]*platform.Platform, error) {
	return nil, fmt.Errorf("platform list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) GetSize() (int64, error) {
	return 0, fmt.Errorf("GetSize is not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) GetSize() (int64, error) {
	return 0, fmt.Errorf("GetSize is not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) MarshalJSON() ([]byte, error) {
	if !m.manifSet {
		return []byte{}, errs.ErrManifestNotSet
	}

	if len(m.rawBody) > 0 {
		return m.rawBody, nil
	}

	return json.Marshal((m.Manifest))
}

func (m *docker1SignedManifest) MarshalJSON() ([]byte, error) {
	if !m.manifSet {
		return []byte{}, errs.ErrManifestNotSet
	}

	return m.SignedManifest.MarshalJSON()
}

func (m *docker1Manifest) MarshalPretty() ([]byte, error) {
	if m == nil {
		return []byte{}, nil
	}
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	if m.r.Reference != "" {
		fmt.Fprintf(tw, "Name:\t%s\n", m.r.Reference)
	}
	fmt.Fprintf(tw, "MediaType:\t%s\n", m.desc.MediaType)
	fmt.Fprintf(tw, "Digest:\t%s\n", m.desc.Digest.String())
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Layers:\t\n")
	for _, d := range m.FSLayers {
		fmt.Fprintf(tw, "  Digest:\t%s\n", string(d.BlobSum))
	}
	err := tw.Flush()
	return buf.Bytes(), err
}

func (m *docker1SignedManifest) MarshalPretty() ([]byte, error) {
	if m == nil {
		return []byte{}, nil
	}
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	if m.r.Reference != "" {
		fmt.Fprintf(tw, "Name:\t%s\n", m.r.Reference)
	}
	fmt.Fprintf(tw, "MediaType:\t%s\n", m.desc.MediaType)
	fmt.Fprintf(tw, "Digest:\t%s\n", m.desc.Digest.String())
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Layers:\t\n")
	for _, d := range m.FSLayers {
		fmt.Fprintf(tw, "  Digest:\t%s\n", string(d.BlobSum))
	}
	err := tw.Flush()
	return buf.Bytes(), err
}

func (m *docker1Manifest) SetConfig(d descriptor.Descriptor) error {
	return fmt.Errorf("set methods not supported for for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) SetConfig(d descriptor.Descriptor) error {
	return fmt.Errorf("set methods not supported for for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) SetLayers(dl []descriptor.Descriptor) error {
	return fmt.Errorf("set methods not supported for for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1SignedManifest) SetLayers(dl []descriptor.Descriptor) error {
	return fmt.Errorf("set methods not supported for for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker1Manifest) SetOrig(origIn any) error {
	orig, ok := origIn.(schema1.Manifest)
	if !ok {
		return errs.ErrUnsupportedMediaType
	}
	if orig.MediaType != mediatype.Docker1Manifest {
		// TODO: error?
		orig.MediaType = mediatype.Docker1Manifest
	}
	mj, err := json.Marshal(orig)
	if err != nil {
		return err
	}
	m.manifSet = true
	m.rawBody = mj
	m.desc = descriptor.Descriptor{
		MediaType: mediatype.Docker1Manifest,
		Digest:    m.desc.DigestAlgo().FromBytes(mj),
		Size:      int64(len(mj)),
	}
	m.Manifest = orig

	return nil
}

func (m *docker1SignedManifest) SetOrig(origIn any) error {
	orig, ok := origIn.(schema1.SignedManifest)
	if !ok {
		return errs.ErrUnsupportedMediaType
	}
	if orig.MediaType != mediatype.Docker1ManifestSigned {
		// TODO: error?
		orig.MediaType = mediatype.Docker1ManifestSigned
	}
	mj, err := json.Marshal(orig)
	if err != nil {
		return err
	}
	m.manifSet = true
	m.rawBody = mj
	m.desc = descriptor.Descriptor{
		MediaType: mediatype.Docker1ManifestSigned,
		Digest:    m.desc.DigestAlgo().FromBytes(orig.Canonical),
		Size:      int64(len(orig.Canonical)),
	}
	m.SignedManifest = orig

	return nil
}
