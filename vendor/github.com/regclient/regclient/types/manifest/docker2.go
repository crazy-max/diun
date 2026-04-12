package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	digest "github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/internal/units"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker/schema2"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/platform"
)

const (
	// MediaTypeDocker2Manifest is the media type when pulling manifests from a v2 registry
	MediaTypeDocker2Manifest = mediatype.Docker2Manifest
	// MediaTypeDocker2ManifestList is the media type when pulling a manifest list from a v2 registry
	MediaTypeDocker2ManifestList = mediatype.Docker2ManifestList
)

type docker2Manifest struct {
	common
	schema2.Manifest
}
type docker2ManifestList struct {
	common
	schema2.ManifestList
}

func (m *docker2Manifest) GetAnnotations() (map[string]string, error) {
	if !m.manifSet {
		return nil, errs.ErrManifestNotSet
	}
	return m.Annotations, nil
}

func (m *docker2Manifest) GetConfig() (descriptor.Descriptor, error) {
	if !m.manifSet {
		return descriptor.Descriptor{}, errs.ErrManifestNotSet
	}
	return m.Config, nil
}

func (m *docker2Manifest) GetConfigDigest() (digest.Digest, error) {
	if !m.manifSet {
		return digest.Digest(""), errs.ErrManifestNotSet
	}
	return m.Config.Digest, nil
}

func (m *docker2ManifestList) GetAnnotations() (map[string]string, error) {
	if !m.manifSet {
		return nil, errs.ErrManifestNotSet
	}
	return m.Annotations, nil
}

func (m *docker2ManifestList) GetConfig() (descriptor.Descriptor, error) {
	return descriptor.Descriptor{}, fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2ManifestList) GetConfigDigest() (digest.Digest, error) {
	return "", fmt.Errorf("config digest not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2Manifest) GetManifestList() ([]descriptor.Descriptor, error) {
	return []descriptor.Descriptor{}, fmt.Errorf("platform descriptor list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2ManifestList) GetManifestList() ([]descriptor.Descriptor, error) {
	if !m.manifSet {
		return []descriptor.Descriptor{}, errs.ErrManifestNotSet
	}
	return m.Manifests, nil
}

func (m *docker2Manifest) GetLayers() ([]descriptor.Descriptor, error) {
	if !m.manifSet {
		return []descriptor.Descriptor{}, errs.ErrManifestNotSet
	}
	return m.Layers, nil
}

func (m *docker2ManifestList) GetLayers() ([]descriptor.Descriptor, error) {
	return []descriptor.Descriptor{}, fmt.Errorf("layers are not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2Manifest) GetOrig() any {
	return m.Manifest
}

func (m *docker2ManifestList) GetOrig() any {
	return m.ManifestList
}

func (m *docker2Manifest) GetPlatformDesc(p *platform.Platform) (*descriptor.Descriptor, error) {
	return nil, fmt.Errorf("platform lookup not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2ManifestList) GetPlatformDesc(p *platform.Platform) (*descriptor.Descriptor, error) {
	if !m.manifSet {
		return nil, errs.ErrManifestNotSet
	}
	if p == nil {
		return nil, fmt.Errorf("invalid input, platform is nil%.0w", errs.ErrNotFound)
	}
	d, err := descriptor.DescriptorListSearch(m.Manifests, descriptor.MatchOpt{Platform: p})
	if err != nil {
		return nil, fmt.Errorf("platform not found: %s%.0w", *p, err)
	}
	return &d, nil
}

func (m *docker2Manifest) GetPlatformList() ([]*platform.Platform, error) {
	return nil, fmt.Errorf("platform list not available for media type %s%.0w", m.desc.MediaType, errs.ErrUnsupportedMediaType)
}

func (m *docker2ManifestList) GetPlatformList() ([]*platform.Platform, error) {
	dl, err := m.GetManifestList()
	if err != nil {
		return nil, err
	}
	return getPlatformList(dl)
}

// GetSize returns the size in bytes of all layers
func (m *docker2Manifest) GetSize() (int64, error) {
	if !m.manifSet {
		return 0, errs.ErrManifestNotSet
	}
	var total int64
	for _, d := range m.Layers {
		total += d.Size
	}
	return total, nil
}

func (m *docker2Manifest) MarshalJSON() ([]byte, error) {
	if !m.manifSet {
		return []byte{}, errs.ErrManifestNotSet
	}
	if len(m.rawBody) > 0 {
		return m.rawBody, nil
	}
	return json.Marshal((m.Manifest))
}

func (m *docker2ManifestList) MarshalJSON() ([]byte, error) {
	if !m.manifSet {
		return []byte{}, errs.ErrManifestNotSet
	}
	if len(m.rawBody) > 0 {
		return m.rawBody, nil
	}
	return json.Marshal((m.ManifestList))
}

func (m *docker2Manifest) MarshalPretty() ([]byte, error) {
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
	if len(m.Annotations) > 0 {
		fmt.Fprintf(tw, "Annotations:\t\n")
		keys := make([]string, 0, len(m.Annotations))
		for k := range m.Annotations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			val := m.Annotations[name]
			fmt.Fprintf(tw, "  %s:\t%s\n", name, val)
		}
	}
	var total int64
	for _, d := range m.Layers {
		total += d.Size
	}
	fmt.Fprintf(tw, "Total Size:\t%s\n", units.HumanSize(float64(total)))
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Config:\t\n")
	err := m.Config.MarshalPrettyTW(tw, "  ")
	if err != nil {
		return []byte{}, err
	}
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Layers:\t\n")
	for _, d := range m.Layers {
		fmt.Fprintf(tw, "\t\n")
		err := d.MarshalPrettyTW(tw, "  ")
		if err != nil {
			return []byte{}, err
		}
	}
	err = tw.Flush()
	return buf.Bytes(), err
}

func (m *docker2ManifestList) MarshalPretty() ([]byte, error) {
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
	if len(m.Annotations) > 0 {
		fmt.Fprintf(tw, "Annotations:\t\n")
		keys := make([]string, 0, len(m.Annotations))
		for k := range m.Annotations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			val := m.Annotations[name]
			fmt.Fprintf(tw, "  %s:\t%s\n", name, val)
		}
	}
	fmt.Fprintf(tw, "\t\n")
	fmt.Fprintf(tw, "Manifests:\t\n")
	for _, d := range m.Manifests {
		fmt.Fprintf(tw, "\t\n")
		dRef := m.r
		if dRef.Reference != "" {
			dRef = dRef.AddDigest(d.Digest.String())
			fmt.Fprintf(tw, "  Name:\t%s\n", dRef.CommonName())
		}
		err := d.MarshalPrettyTW(tw, "  ")
		if err != nil {
			return []byte{}, err
		}
	}
	err := tw.Flush()
	return buf.Bytes(), err
}

func (m *docker2Manifest) SetAnnotation(key, val string) error {
	if !m.manifSet {
		return errs.ErrManifestNotSet
	}
	if m.Annotations == nil {
		m.Annotations = map[string]string{}
	}
	if val != "" {
		m.Annotations[key] = val
	} else {
		delete(m.Annotations, key)
	}
	return m.updateDesc()
}

func (m *docker2ManifestList) SetAnnotation(key, val string) error {
	if !m.manifSet {
		return errs.ErrManifestNotSet
	}
	if m.Annotations == nil {
		m.Annotations = map[string]string{}
	}
	if val != "" {
		m.Annotations[key] = val
	} else {
		delete(m.Annotations, key)
	}
	return m.updateDesc()
}

func (m *docker2Manifest) SetConfig(d descriptor.Descriptor) error {
	if !m.manifSet {
		return errs.ErrManifestNotSet
	}
	m.Config = d
	return m.updateDesc()
}

func (m *docker2Manifest) SetLayers(dl []descriptor.Descriptor) error {
	if !m.manifSet {
		return errs.ErrManifestNotSet
	}
	m.Layers = dl
	return m.updateDesc()
}

func (m *docker2ManifestList) SetManifestList(dl []descriptor.Descriptor) error {
	if !m.manifSet {
		return errs.ErrManifestNotSet
	}
	m.Manifests = dl
	return m.updateDesc()
}

func (m *docker2Manifest) SetOrig(origIn any) error {
	orig, ok := origIn.(schema2.Manifest)
	if !ok {
		return errs.ErrUnsupportedMediaType
	}
	if orig.MediaType != mediatype.Docker2Manifest {
		// TODO: error?
		orig.MediaType = mediatype.Docker2Manifest
	}
	m.manifSet = true
	m.Manifest = orig
	return m.updateDesc()
}

func (m *docker2ManifestList) SetOrig(origIn any) error {
	orig, ok := origIn.(schema2.ManifestList)
	if !ok {
		return errs.ErrUnsupportedMediaType
	}
	if orig.MediaType != mediatype.Docker2ManifestList {
		// TODO: error?
		orig.MediaType = mediatype.Docker2ManifestList
	}
	m.manifSet = true
	m.ManifestList = orig
	return m.updateDesc()
}

func (m *docker2Manifest) updateDesc() error {
	mj, err := json.Marshal(m.Manifest)
	if err != nil {
		return err
	}
	m.rawBody = mj
	m.desc = descriptor.Descriptor{
		MediaType: mediatype.Docker2Manifest,
		Digest:    m.desc.DigestAlgo().FromBytes(mj),
		Size:      int64(len(mj)),
	}
	return nil
}

func (m *docker2ManifestList) updateDesc() error {
	mj, err := json.Marshal(m.ManifestList)
	if err != nil {
		return err
	}
	m.rawBody = mj
	m.desc = descriptor.Descriptor{
		MediaType: mediatype.Docker2ManifestList,
		Digest:    m.desc.DigestAlgo().FromBytes(mj),
		Size:      int64(len(mj)),
	}
	return nil
}
