// Package blob is the underlying type for pushing and pulling blobs.
package blob

import (
	"io"
	"net/http"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types/descriptor"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

// Blob interface is used for returning blobs.
type Blob interface {
	// GetDescriptor returns the descriptor associated with the blob.
	GetDescriptor() descriptor.Descriptor
	// RawBody returns the raw content of the blob.
	RawBody() ([]byte, error)
	// RawHeaders returns the headers received from the registry.
	RawHeaders() http.Header
	// Response returns the response associated with the blob.
	Response() *http.Response

	// Digest returns the provided or calculated digest of the blob.
	//
	// Deprecated: Digest should be replaced by GetDescriptor().Digest.
	Digest() digest.Digest
	// Length returns the provided or calculated length of the blob.
	//
	// Deprecated: Length should be replaced by GetDescriptor().Size.
	Length() int64
	// MediaType returns the Content-Type header received from the registry.
	//
	// Deprecated: MediaType should be replaced by GetDescriptor().MediaType.
	MediaType() string
}

type blobConfig struct {
	desc    descriptor.Descriptor
	header  http.Header
	image   *v1.Image
	r       ref.Ref
	rdr     io.Reader
	resp    *http.Response
	rawBody []byte
}

// Opts is used for options to create a new blob.
type Opts func(*blobConfig)

// WithDesc specifies the descriptor associated with the blob.
func WithDesc(d descriptor.Descriptor) Opts {
	return func(bc *blobConfig) {
		bc.desc = d
	}
}

// WithHeader defines the headers received when pulling a blob.
func WithHeader(header http.Header) Opts {
	return func(bc *blobConfig) {
		bc.header = header
	}
}

// WithImage provides the OCI Image config needed for config blobs.
func WithImage(image v1.Image) Opts {
	return func(bc *blobConfig) {
		bc.image = &image
	}
}

// WithRawBody defines the raw blob contents for OCIConfig.
func WithRawBody(raw []byte) Opts {
	return func(bc *blobConfig) {
		bc.rawBody = raw
	}
}

// WithReader defines the reader for a new blob.
func WithReader(rc io.Reader) Opts {
	return func(bc *blobConfig) {
		bc.rdr = rc
	}
}

// WithRef specifies the reference where the blob was pulled from.
func WithRef(r ref.Ref) Opts {
	return func(bc *blobConfig) {
		bc.r = r
	}
}

// WithResp includes the http response, which is used to extract the headers and reader.
func WithResp(resp *http.Response) Opts {
	return func(bc *blobConfig) {
		bc.resp = resp
		if bc.header == nil && resp != nil {
			bc.header = resp.Header
		}
	}
}
