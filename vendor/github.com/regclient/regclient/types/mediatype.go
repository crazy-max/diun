package types

import (
	"github.com/regclient/regclient/types/mediatype"
)

const (
	// MediaTypeDocker1Manifest deprecated media type for docker schema1 manifests.
	//
	// Deprecated: replace with [mediatype.Docker1Manifest].
	//go:fix inline
	MediaTypeDocker1Manifest = mediatype.Docker1Manifest
	// MediaTypeDocker1ManifestSigned is a deprecated schema1 manifest with jws signing.
	//
	// Deprecated: replace with [mediatype.Docker1ManifestSigned].
	//go:fix inline
	MediaTypeDocker1ManifestSigned = mediatype.Docker1ManifestSigned
	// MediaTypeDocker2Manifest is the media type when pulling manifests from a v2 registry.
	//
	// Deprecated: replace with [mediatype.Docker2Manifest].
	//go:fix inline
	MediaTypeDocker2Manifest = mediatype.Docker2Manifest
	// MediaTypeDocker2ManifestList is the media type when pulling a manifest list from a v2 registry.
	//
	// Deprecated: replace with [mediatype.Docker2ManifestList].
	//go:fix inline
	MediaTypeDocker2ManifestList = mediatype.Docker2ManifestList
	// MediaTypeDocker2ImageConfig is for the configuration json object media type.
	//
	// Deprecated: replace with [mediatype.Docker2ImageConfig].
	//go:fix inline
	MediaTypeDocker2ImageConfig = mediatype.Docker2ImageConfig
	// MediaTypeOCI1Artifact EXPERIMENTAL OCI v1 artifact media type.
	//
	// Deprecated: replace with [mediatype.OCI1Artifact].
	//go:fix inline
	MediaTypeOCI1Artifact = mediatype.OCI1Artifact
	// MediaTypeOCI1Manifest OCI v1 manifest media type.
	//
	// Deprecated: replace with [mediatype.OCI1Manifest].
	//go:fix inline
	MediaTypeOCI1Manifest = mediatype.OCI1Manifest
	// MediaTypeOCI1ManifestList OCI v1 manifest list media type.
	//
	// Deprecated: replace with [mediatype.OCI1ManifestList].
	//go:fix inline
	MediaTypeOCI1ManifestList = mediatype.OCI1ManifestList
	// MediaTypeOCI1ImageConfig OCI v1 configuration json object media type.
	//
	// Deprecated: replace with [mediatype.OCI1ImageConfig].
	//go:fix inline
	MediaTypeOCI1ImageConfig = mediatype.OCI1ImageConfig
	// MediaTypeDocker2LayerGzip is the default compressed layer for docker schema2.
	//
	// Deprecated: replace with [mediatype.Docker2LayerGzip].
	//go:fix inline
	MediaTypeDocker2LayerGzip = mediatype.Docker2LayerGzip
	// MediaTypeDocker2ForeignLayer is the default compressed layer for foreign layers in docker schema2.
	//
	// Deprecated: replace with [mediatype.Docker2ForeignLayer].
	//go:fix inline
	MediaTypeDocker2ForeignLayer = mediatype.Docker2ForeignLayer
	// MediaTypeOCI1Layer is the uncompressed layer for OCIv1.
	//
	// Deprecated: replace with [mediatype.OCI1Layer].
	//go:fix inline
	MediaTypeOCI1Layer = mediatype.OCI1Layer
	// MediaTypeOCI1LayerGzip is the gzip compressed layer for OCI v1.
	//
	// Deprecated: replace with [mediatype.OCI1LayerGzip].
	//go:fix inline
	MediaTypeOCI1LayerGzip = mediatype.OCI1LayerGzip
	// MediaTypeOCI1LayerZstd is the zstd compressed layer for OCI v1.
	//
	// Deprecated: replace with [mediatype.OCI1LayerZstd].
	//go:fix inline
	MediaTypeOCI1LayerZstd = mediatype.OCI1LayerZstd
	// MediaTypeOCI1ForeignLayer is the foreign layer for OCI v1.
	//
	// Deprecated: replace with [mediatype.OCI1ForeignLayer].
	//go:fix inline
	MediaTypeOCI1ForeignLayer = mediatype.OCI1ForeignLayer
	// MediaTypeOCI1ForeignLayerGzip is the gzip compressed foreign layer for OCI v1.
	//
	// Deprecated: replace with [mediatype.OCI1ForeignLayerGzip].
	//go:fix inline
	MediaTypeOCI1ForeignLayerGzip = mediatype.OCI1ForeignLayerGzip
	// MediaTypeOCI1ForeignLayerZstd is the zstd compressed foreign layer for OCI v1.
	//
	// Deprecated: replace with [mediatype.OCI1ForeignLayerZstd].
	//go:fix inline
	MediaTypeOCI1ForeignLayerZstd = mediatype.OCI1ForeignLayerZstd
	// MediaTypeOCI1Empty is used for blobs containing the empty JSON data `{}`.
	//
	// Deprecated: replace with [mediatype.OCI1Empty].
	//go:fix inline
	MediaTypeOCI1Empty = mediatype.OCI1Empty
	// MediaTypeBuildkitCacheConfig is used by buildkit cache images.
	//
	// Deprecated: replace with [mediatype.BuildkitCacheConfig].
	//go:fix inline
	MediaTypeBuildkitCacheConfig = mediatype.BuildkitCacheConfig
)

// MediaTypeBase cleans the Content-Type header to return only the lower case base media type.
//
// Deprecated: replace with [mediatype.Base].
//
//go:fix inline
var MediaTypeBase = mediatype.Base
