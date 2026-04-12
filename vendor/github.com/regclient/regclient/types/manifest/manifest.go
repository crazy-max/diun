// Package manifest abstracts the various types of supported manifests.
// Supported types include OCI index and image, and Docker manifest list and manifest.
package manifest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	// Crypto libraries are included for go-digest.
	_ "crypto/sha256"
	_ "crypto/sha512"

	digest "github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker/schema1"
	"github.com/regclient/regclient/types/docker/schema2"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

// Manifest interface is implemented by all supported manifests but
// many calls are only supported by certain underlying media types.
type Manifest interface {
	GetDescriptor() descriptor.Descriptor
	GetOrig() any
	GetRef() ref.Ref
	IsList() bool
	IsSet() bool
	MarshalJSON() ([]byte, error)
	RawBody() ([]byte, error)
	RawHeaders() (http.Header, error)
	SetOrig(any) error

	// Deprecated: GetConfig should be accessed using [Imager] interface.
	GetConfig() (descriptor.Descriptor, error)
	// Deprecated: GetLayers should be accessed using [Imager] interface.
	GetLayers() ([]descriptor.Descriptor, error)

	// Deprecated: GetManifestList should be accessed using [Indexer] interface.
	GetManifestList() ([]descriptor.Descriptor, error)

	// Deprecated: GetConfigDigest should be replaced with [GetConfig].
	GetConfigDigest() (digest.Digest, error)
	// Deprecated: GetDigest should be replaced with GetDescriptor().Digest, see [GetDescriptor].
	GetDigest() digest.Digest
	// Deprecated: GetMediaType should be replaced with GetDescriptor().MediaType, see [GetDescriptor].
	GetMediaType() string
	// Deprecated: GetPlatformDesc method should be replaced with [manifest.GetPlatformDesc].
	GetPlatformDesc(p *platform.Platform) (*descriptor.Descriptor, error)
	// Deprecated: GetPlatformList method should be replaced with [manifest.GetPlatformList].
	GetPlatformList() ([]*platform.Platform, error)
	// Deprecated: GetRateLimit method should be replaced with [manifest.GetRateLimit].
	GetRateLimit() types.RateLimit
	// Deprecated: HasRateLimit method should be replaced with [manifest.HasRateLimit].
	HasRateLimit() bool
}

// Annotator is used by manifests that support annotations.
// Note this will work for Docker manifests despite the spec not officially supporting it.
type Annotator interface {
	GetAnnotations() (map[string]string, error)
	SetAnnotation(key, val string) error
}

// Indexer is used by manifests that contain a manifest list.
type Indexer interface {
	GetManifestList() ([]descriptor.Descriptor, error)
	SetManifestList(dl []descriptor.Descriptor) error
}

// Imager is used by manifests packaging an image.
type Imager interface {
	GetConfig() (descriptor.Descriptor, error)
	GetLayers() ([]descriptor.Descriptor, error)
	SetConfig(d descriptor.Descriptor) error
	SetLayers(dl []descriptor.Descriptor) error
	GetSize() (int64, error)
}

// Subjecter is used by manifests that may have a subject field.
type Subjecter interface {
	GetSubject() (*descriptor.Descriptor, error)
	SetSubject(d *descriptor.Descriptor) error
}

type manifestConfig struct {
	r      ref.Ref
	desc   descriptor.Descriptor
	raw    []byte
	orig   any
	header http.Header
}
type Opts func(*manifestConfig)

// New creates a new manifest based on provided options.
// The digest for the manifest will be checked against the descriptor, reference, or headers, depending on which is available first (later digests will be ignored).
func New(opts ...Opts) (Manifest, error) {
	mc := manifestConfig{}
	for _, opt := range opts {
		opt(&mc)
	}
	c := common{
		r:         mc.r,
		desc:      mc.desc,
		rawBody:   mc.raw,
		rawHeader: mc.header,
	}
	if c.r.Digest != "" && c.desc.Digest == "" {
		dig, err := digest.Parse(c.r.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse digest from ref: %w", err)
		}
		c.desc.Digest = dig
	}
	// extract fields from header where available
	if mc.header != nil {
		if c.desc.MediaType == "" {
			c.desc.MediaType = mediatype.Base(mc.header.Get("Content-Type"))
		}
		if c.desc.Size == 0 {
			cl, _ := strconv.Atoi(mc.header.Get("Content-Length"))
			c.desc.Size = int64(cl)
		}
		if c.desc.Digest == "" {
			c.desc.Digest, _ = digest.Parse(mc.header.Get("Docker-Content-Digest"))
		}
		c.setRateLimit(mc.header)
	}
	if mc.orig != nil {
		return fromOrig(c, mc.orig)
	}
	return fromCommon(c)
}

// WithDesc specifies the descriptor for the manifest.
func WithDesc(desc descriptor.Descriptor) Opts {
	return func(mc *manifestConfig) {
		mc.desc = desc
	}
}

// WithHeader provides the headers from the response when pulling the manifest.
func WithHeader(header http.Header) Opts {
	return func(mc *manifestConfig) {
		mc.header = header
	}
}

// WithOrig provides the original manifest variable.
func WithOrig(orig any) Opts {
	return func(mc *manifestConfig) {
		mc.orig = orig
	}
}

// WithRaw provides the manifest bytes or HTTP response body.
func WithRaw(raw []byte) Opts {
	return func(mc *manifestConfig) {
		mc.raw = raw
	}
}

// WithRef provides the reference used to get the manifest.
func WithRef(r ref.Ref) Opts {
	return func(mc *manifestConfig) {
		mc.r = r
	}
}

// GetDigest returns the digest from the manifest descriptor.
func GetDigest(m Manifest) digest.Digest {
	d := m.GetDescriptor()
	return d.Digest
}

// GetMediaType returns the media type from the manifest descriptor.
func GetMediaType(m Manifest) string {
	d := m.GetDescriptor()
	return d.MediaType
}

// GetPlatformDesc returns the descriptor for a specific platform from an index.
func GetPlatformDesc(m Manifest, p *platform.Platform) (*descriptor.Descriptor, error) {
	if p == nil {
		return nil, fmt.Errorf("invalid input, platform is nil%.0w", errs.ErrNotFound)
	}
	mi, ok := m.(Indexer)
	if !ok {
		return nil, fmt.Errorf("unsupported manifest type: %s", m.GetDescriptor().MediaType)
	}
	dl, err := mi.GetManifestList()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest list: %w", err)
	}
	d, err := descriptor.DescriptorListSearch(dl, descriptor.MatchOpt{Platform: p})
	if err != nil {
		return nil, fmt.Errorf("platform not found: %s%.0w", *p, err)
	}
	return &d, nil
}

// GetPlatformList returns the list of platforms from an index.
func GetPlatformList(m Manifest) ([]*platform.Platform, error) {
	mi, ok := m.(Indexer)
	if !ok {
		return nil, fmt.Errorf("unsupported manifest type: %s", m.GetDescriptor().MediaType)
	}
	dl, err := mi.GetManifestList()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest list: %w", err)
	}
	return getPlatformList(dl)
}

// GetRateLimit returns the current rate limit seen in headers.
func GetRateLimit(m Manifest) types.RateLimit {
	rl := types.RateLimit{}
	header, err := m.RawHeaders()
	if err != nil {
		return rl
	}
	// check for rate limit headers
	rlLimit := header.Get("RateLimit-Limit")
	rlRemain := header.Get("RateLimit-Remaining")
	rlReset := header.Get("RateLimit-Reset")
	if rlLimit != "" {
		lpSplit := strings.Split(rlLimit, ",")
		lSplit := strings.Split(lpSplit[0], ";")
		rlLimitI, err := strconv.Atoi(lSplit[0])
		if err != nil {
			rl.Limit = 0
		} else {
			rl.Limit = rlLimitI
		}
		if len(lSplit) > 1 {
			rl.Policies = lpSplit
		} else if len(lpSplit) > 1 {
			rl.Policies = lpSplit[1:]
		}
	}
	if rlRemain != "" {
		rSplit := strings.Split(rlRemain, ";")
		rlRemainI, err := strconv.Atoi(rSplit[0])
		if err != nil {
			rl.Remain = 0
		} else {
			rl.Remain = rlRemainI
			rl.Set = true
		}
	}
	if rlReset != "" {
		rlResetI, err := strconv.Atoi(rlReset)
		if err != nil {
			rl.Reset = 0
		} else {
			rl.Reset = rlResetI
		}
	}
	return rl
}

// HasRateLimit indicates whether the rate limit is set and available.
func HasRateLimit(m Manifest) bool {
	rl := GetRateLimit(m)
	return rl.Set
}

// OCIIndexFromAny converts manifest lists to an OCI index.
func OCIIndexFromAny(orig any) (v1.Index, error) {
	ociI := v1.Index{
		Versioned: v1.IndexSchemaVersion,
		MediaType: mediatype.OCI1ManifestList,
	}
	switch orig := orig.(type) {
	case schema2.ManifestList:
		ociI.Manifests = orig.Manifests
		ociI.Annotations = orig.Annotations
	case v1.Index:
		ociI = orig
	default:
		return ociI, fmt.Errorf("unable to convert %T to OCI index", orig)
	}
	return ociI, nil
}

// OCIIndexToAny converts from an OCI index back to the manifest list.
func OCIIndexToAny(ociI v1.Index, origP any) error {
	// reflect is used to handle both *interface and *Manifest
	rv := reflect.ValueOf(origP)
	for rv.IsValid() && rv.Type().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return fmt.Errorf("invalid manifest output parameter: %T", origP)
	}
	if !rv.CanSet() {
		return fmt.Errorf("manifest output must be a pointer: %T", origP)
	}
	origR := rv.Interface()
	switch orig := (origR).(type) {
	case schema2.ManifestList:
		orig.Versioned = schema2.ManifestListSchemaVersion
		orig.Manifests = ociI.Manifests
		orig.Annotations = ociI.Annotations
		rv.Set(reflect.ValueOf(orig))
	case v1.Index:
		rv.Set(reflect.ValueOf(ociI))
	default:
		return fmt.Errorf("unable to convert OCI index to %T", origR)
	}
	return nil
}

// OCIManifestFromAny converts an image manifest to an OCI manifest.
func OCIManifestFromAny(orig any) (v1.Manifest, error) {
	ociM := v1.Manifest{
		Versioned: v1.ManifestSchemaVersion,
		MediaType: mediatype.OCI1Manifest,
	}
	switch orig := orig.(type) {
	case schema2.Manifest:
		ociM.Config = orig.Config
		ociM.Layers = orig.Layers
		ociM.Annotations = orig.Annotations
	case v1.Manifest:
		ociM = orig
	default:
		// TODO: consider supporting Docker schema v1 media types
		return ociM, fmt.Errorf("unable to convert %T to OCI image", orig)
	}
	return ociM, nil
}

// OCIManifestToAny converts an OCI manifest back to the image manifest.
func OCIManifestToAny(ociM v1.Manifest, origP any) error {
	// reflect is used to handle both *interface and *Manifest
	rv := reflect.ValueOf(origP)
	for rv.IsValid() && rv.Type().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return fmt.Errorf("invalid manifest output parameter: %T", origP)
	}
	if !rv.CanSet() {
		return fmt.Errorf("manifest output must be a pointer: %T", origP)
	}
	origR := rv.Interface()
	switch orig := (origR).(type) {
	case schema2.Manifest:
		orig.Versioned = schema2.ManifestSchemaVersion
		orig.Config = ociM.Config
		orig.Layers = ociM.Layers
		orig.Annotations = ociM.Annotations
		rv.Set(reflect.ValueOf(orig))
	case v1.Manifest:
		rv.Set(reflect.ValueOf(ociM))
	default:
		// Docker schema v1 will not be supported, can't resign, and no need for unsigned
		return fmt.Errorf("unable to convert OCI image to %T", origR)
	}
	return nil
}

// FromOrig creates a new manifest from the original upstream manifest type.
// This method should be used if you are creating a new manifest rather than pulling one from a registry.
func fromOrig(c common, orig any) (Manifest, error) {
	var mt string
	var m Manifest
	origDigest := c.desc.Digest

	mj, err := json.Marshal(orig)
	if err != nil {
		return nil, err
	}
	c.manifSet = true
	if len(c.rawBody) == 0 {
		c.rawBody = mj
	}
	if _, ok := orig.(schema1.SignedManifest); !ok {
		c.desc.Digest = c.desc.DigestAlgo().FromBytes(mj)
	}
	if c.desc.Size == 0 {
		c.desc.Size = int64(len(mj))
	}
	// create manifest based on type
	switch mOrig := orig.(type) {
	case schema1.Manifest:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.Docker1Manifest
		m = &docker1Manifest{
			common:   c,
			Manifest: mOrig,
		}
	case schema1.SignedManifest:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.Docker1ManifestSigned
		// recompute digest on the canonical data
		c.desc.Digest = c.desc.DigestAlgo().FromBytes(mOrig.Canonical)
		m = &docker1SignedManifest{
			common:         c,
			SignedManifest: mOrig,
		}
	case schema2.Manifest:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.Docker2Manifest
		m = &docker2Manifest{
			common:   c,
			Manifest: mOrig,
		}
	case schema2.ManifestList:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.Docker2ManifestList
		m = &docker2ManifestList{
			common:       c,
			ManifestList: mOrig,
		}
	case v1.Manifest:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.OCI1Manifest
		m = &oci1Manifest{
			common:   c,
			Manifest: mOrig,
		}
	case v1.Index:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.OCI1ManifestList
		m = &oci1Index{
			common: c,
			Index:  orig.(v1.Index),
		}
	case v1.ArtifactManifest:
		mt = mOrig.MediaType
		c.desc.MediaType = mediatype.OCI1Artifact
		m = &oci1Artifact{
			common:           c,
			ArtifactManifest: mOrig,
		}
	default:
		return nil, fmt.Errorf("unsupported type to convert to a manifest: %T", orig)
	}
	// verify media type
	err = verifyMT(c.desc.MediaType, mt)
	if err != nil {
		return nil, err
	}
	// verify digest didn't change
	if origDigest != "" && origDigest != c.desc.Digest {
		return nil, fmt.Errorf("manifest digest mismatch, expected %s, computed %s%.0w", origDigest, c.desc.Digest, errs.ErrDigestMismatch)
	}
	return m, nil
}

// fromCommon is used to create a manifest when the underlying manifest struct is not provided.
func fromCommon(c common) (Manifest, error) {
	var err error
	var m Manifest
	var mt string
	origDigest := c.desc.Digest
	// extract common data from from rawBody
	if len(c.rawBody) > 0 {
		c.manifSet = true
		// extract media type from body, either explicitly or with duck typing
		if c.desc.MediaType == "" {
			mt := struct {
				MediaType     string                  `json:"mediaType,omitempty"`
				SchemaVersion int                     `json:"schemaVersion,omitempty"`
				Signatures    []any                   `json:"signatures,omitempty"`
				Manifests     []descriptor.Descriptor `json:"manifests,omitempty"`
				Layers        []descriptor.Descriptor `json:"layers,omitempty"`
			}{}
			err = json.Unmarshal(c.rawBody, &mt)
			if mt.MediaType != "" {
				c.desc.MediaType = mt.MediaType
			} else if mt.SchemaVersion == 1 && len(mt.Signatures) > 0 {
				c.desc.MediaType = mediatype.Docker1ManifestSigned
			} else if mt.SchemaVersion == 1 {
				c.desc.MediaType = mediatype.Docker1Manifest
			} else if len(mt.Manifests) > 0 {
				if strings.HasPrefix(mt.Manifests[0].MediaType, "application/vnd.docker.") {
					c.desc.MediaType = mediatype.Docker2ManifestList
				} else {
					c.desc.MediaType = mediatype.OCI1ManifestList
				}
			} else if len(mt.Layers) > 0 {
				if strings.HasPrefix(mt.Layers[0].MediaType, "application/vnd.docker.") {
					c.desc.MediaType = mediatype.Docker2Manifest
				} else {
					c.desc.MediaType = mediatype.OCI1Manifest
				}
			}
		}
		// compute digest
		if c.desc.MediaType != mediatype.Docker1ManifestSigned {
			d := c.desc.DigestAlgo().FromBytes(c.rawBody)
			c.desc.Digest = d
			c.desc.Size = int64(len(c.rawBody))
		}
	}
	switch c.desc.MediaType {
	case mediatype.Docker1Manifest:
		var mOrig schema1.Manifest
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &docker1Manifest{common: c, Manifest: mOrig}
	case mediatype.Docker1ManifestSigned:
		var mOrig schema1.SignedManifest
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
			d := c.desc.DigestAlgo().FromBytes(mOrig.Canonical)
			c.desc.Digest = d
			c.desc.Size = int64(len(mOrig.Canonical))
		}
		m = &docker1SignedManifest{common: c, SignedManifest: mOrig}
	case mediatype.Docker2Manifest:
		var mOrig schema2.Manifest
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &docker2Manifest{common: c, Manifest: mOrig}
	case mediatype.Docker2ManifestList:
		var mOrig schema2.ManifestList
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &docker2ManifestList{common: c, ManifestList: mOrig}
	case mediatype.OCI1Manifest:
		var mOrig v1.Manifest
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &oci1Manifest{common: c, Manifest: mOrig}
	case mediatype.OCI1ManifestList:
		var mOrig v1.Index
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &oci1Index{common: c, Index: mOrig}
	case mediatype.OCI1Artifact:
		var mOrig v1.ArtifactManifest
		if len(c.rawBody) > 0 {
			err = json.Unmarshal(c.rawBody, &mOrig)
			mt = mOrig.MediaType
		}
		m = &oci1Artifact{common: c, ArtifactManifest: mOrig}
	default:
		return nil, fmt.Errorf("%w: \"%s\"", errs.ErrUnsupportedMediaType, c.desc.MediaType)
	}
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling manifest for %s: %w", c.r.CommonName(), err)
	}
	// verify media type
	err = verifyMT(c.desc.MediaType, mt)
	if err != nil {
		return nil, err
	}
	// verify digest didn't change
	if origDigest != "" && origDigest != c.desc.Digest {
		return nil, fmt.Errorf("manifest digest mismatch, expected %s, computed %s%.0w", origDigest, c.desc.Digest, errs.ErrDigestMismatch)
	}
	return m, nil
}

func verifyMT(expected, received string) error {
	if received != "" && expected != received {
		return fmt.Errorf("manifest contains an unexpected media type: expected %s, received %s", expected, received)
	}
	return nil
}

func getPlatformList(dl []descriptor.Descriptor) ([]*platform.Platform, error) {
	var l []*platform.Platform
	for _, d := range dl {
		if d.Platform != nil {
			l = append(l, d.Platform)
		}
	}
	return l, nil
}
