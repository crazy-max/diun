package blob

import (
	"net/http"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/ref"
)

// Common was previously an interface. A type alias is provided for upgrades.
type Common = *BCommon

// BCommon is a common struct for all blobs which includes various shared methods.
type BCommon struct {
	r         ref.Ref
	desc      descriptor.Descriptor
	blobSet   bool
	rawHeader http.Header
	resp      *http.Response
}

// GetDescriptor returns the descriptor associated with the blob.
func (c *BCommon) GetDescriptor() descriptor.Descriptor {
	return c.desc
}

// Digest returns the provided or calculated digest of the blob.
//
// Deprecated: Digest should be replaced by GetDescriptor().Digest, see [GetDescriptor].
//
//go:fix inline
func (c *BCommon) Digest() digest.Digest {
	return c.desc.Digest
}

// Length returns the provided or calculated length of the blob.
//
// Deprecated: Length should be replaced by GetDescriptor().Size, see [GetDescriptor].
//
//go:fix inline
func (c *BCommon) Length() int64 {
	return c.desc.Size
}

// MediaType returns the Content-Type header received from the registry.
//
// Deprecated: MediaType should be replaced by GetDescriptor().MediaType, see [GetDescriptor].
//
//go:fix inline
func (c *BCommon) MediaType() string {
	return c.desc.MediaType
}

// RawHeaders returns the headers received from the registry.
func (c *BCommon) RawHeaders() http.Header {
	return c.rawHeader
}

// Response returns the response associated with the blob.
func (c *BCommon) Response() *http.Response {
	return c.resp
}
