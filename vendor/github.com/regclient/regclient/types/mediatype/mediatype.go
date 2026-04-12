// Package mediatype defines well known media types.
package mediatype

import (
	"regexp"
	"strings"
)

const (
	// Docker1Manifest deprecated media type for docker schema1 manifests.
	Docker1Manifest = "application/vnd.docker.distribution.manifest.v1+json"
	// Docker1ManifestSigned is a deprecated schema1 manifest with jws signing.
	Docker1ManifestSigned = "application/vnd.docker.distribution.manifest.v1+prettyjws"
	// Docker2Manifest is the media type when pulling manifests from a v2 registry.
	Docker2Manifest = "application/vnd.docker.distribution.manifest.v2+json"
	// Docker2ManifestList is the media type when pulling a manifest list from a v2 registry.
	Docker2ManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	// Docker2ImageConfig is for the configuration json object media type.
	Docker2ImageConfig = "application/vnd.docker.container.image.v1+json"
	// OCI1Artifact EXPERIMENTAL OCI v1 artifact media type.
	OCI1Artifact = "application/vnd.oci.artifact.manifest.v1+json"
	// OCI1Manifest OCI v1 manifest media type.
	OCI1Manifest = "application/vnd.oci.image.manifest.v1+json"
	// OCI1ManifestList OCI v1 manifest list media type.
	OCI1ManifestList = "application/vnd.oci.image.index.v1+json"
	// OCI1ImageConfig OCI v1 configuration json object media type.
	OCI1ImageConfig = "application/vnd.oci.image.config.v1+json"
	// Docker2Layer is the uncompressed layer for docker schema2.
	Docker2Layer = "application/vnd.docker.image.rootfs.diff.tar"
	// Docker2LayerGzip is the default compressed layer for docker schema2.
	Docker2LayerGzip = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	// Docker2LayerZstd is the default compressed layer for docker schema2.
	Docker2LayerZstd = "application/vnd.docker.image.rootfs.diff.tar.zstd"
	// Docker2ForeignLayer is the default compressed layer for foreign layers in docker schema2.
	Docker2ForeignLayer = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
	// OCI1Layer is the uncompressed layer for OCIv1.
	OCI1Layer = "application/vnd.oci.image.layer.v1.tar"
	// OCI1LayerGzip is the gzip compressed layer for OCI v1.
	OCI1LayerGzip = "application/vnd.oci.image.layer.v1.tar+gzip"
	// OCI1LayerZstd is the zstd compressed layer for OCI v1.
	OCI1LayerZstd = "application/vnd.oci.image.layer.v1.tar+zstd"
	// OCI1ForeignLayer is the foreign layer for OCI v1.
	OCI1ForeignLayer = "application/vnd.oci.image.layer.nondistributable.v1.tar"
	// OCI1ForeignLayerGzip is the gzip compressed foreign layer for OCI v1.
	OCI1ForeignLayerGzip = "application/vnd.oci.image.layer.nondistributable.v1.tar+gzip"
	// OCI1ForeignLayerZstd is the zstd compressed foreign layer for OCI v1.
	OCI1ForeignLayerZstd = "application/vnd.oci.image.layer.nondistributable.v1.tar+zstd"
	// OCI1Empty is used for blobs containing the empty JSON data `{}`.
	OCI1Empty = "application/vnd.oci.empty.v1+json"
	// BuildkitCacheConfig is used by buildkit cache images.
	BuildkitCacheConfig = "application/vnd.buildkit.cacheconfig.v0"
)

// Base cleans the Content-Type header to return only the lower case base media type.
func Base(orig string) string {
	base, _, _ := strings.Cut(orig, ";")
	return strings.TrimSpace(strings.ToLower(base))
}

// Valid returns true if the media type matches the rfc6838 4.2 naming requirements.
func Valid(mt string) bool {
	return validateRegexp.MatchString(mt)
}

var validateRegexp = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}/[A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}$`)
