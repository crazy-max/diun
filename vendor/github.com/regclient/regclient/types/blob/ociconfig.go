package blob

import (
	"encoding/json"
	"fmt"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
)

// OCIConfig was previously an interface. A type alias is provided for upgrading.
type OCIConfig = *BOCIConfig

// BOCIConfig includes an OCI Image Config struct that may be extracted from or pushed to a blob.
type BOCIConfig struct {
	BCommon
	rawBody []byte
	image   v1.Image
}

// NewOCIConfig creates a new BOCIConfig.
// When created from an existing blob, a BOCIConfig will be created using BReader.ToOCIConfig().
func NewOCIConfig(opts ...Opts) *BOCIConfig {
	bc := blobConfig{}
	for _, opt := range opts {
		opt(&bc)
	}
	if bc.image != nil && len(bc.rawBody) == 0 {
		var err error
		bc.rawBody, err = json.Marshal(bc.image)
		if err != nil {
			bc.rawBody = []byte{}
		}
	}
	if len(bc.rawBody) > 0 {
		if bc.image == nil {
			bc.image = &v1.Image{}
			err := json.Unmarshal(bc.rawBody, bc.image)
			if err != nil {
				bc.image = nil
			}
		}
		// force descriptor to match raw body, even if we generated the raw body
		bc.desc.Digest = bc.desc.DigestAlgo().FromBytes(bc.rawBody)
		bc.desc.Size = int64(len(bc.rawBody))
		if bc.desc.MediaType == "" {
			bc.desc.MediaType = mediatype.OCI1ImageConfig
		}
	}
	b := BOCIConfig{
		BCommon: BCommon{
			desc:      bc.desc,
			r:         bc.r,
			rawHeader: bc.header,
			resp:      bc.resp,
		},
		rawBody: bc.rawBody,
	}
	if bc.image != nil {
		b.image = *bc.image
		b.blobSet = true
	}
	return &b
}

// GetConfig returns OCI config.
func (oc *BOCIConfig) GetConfig() v1.Image {
	return oc.image
}

// RawBody returns the original body from the request.
func (oc *BOCIConfig) RawBody() ([]byte, error) {
	var err error
	if !oc.blobSet {
		return []byte{}, fmt.Errorf("Blob is not defined")
	}
	if len(oc.rawBody) == 0 {
		oc.rawBody, err = json.Marshal(oc.image)
	}
	return oc.rawBody, err
}

// SetConfig updates the config, including raw body and descriptor.
func (oc *BOCIConfig) SetConfig(image v1.Image) {
	oc.image = image
	oc.rawBody, _ = json.Marshal(oc.image)
	if oc.desc.MediaType == "" {
		oc.desc.MediaType = mediatype.OCI1ImageConfig
	}
	oc.desc.Digest = oc.desc.DigestAlgo().FromBytes(oc.rawBody)
	oc.desc.Size = int64(len(oc.rawBody))
	oc.blobSet = true
}

// MarshalJSON passes through the marshalling to the underlying image if rawBody is not available.
func (oc *BOCIConfig) MarshalJSON() ([]byte, error) {
	if !oc.blobSet {
		return []byte{}, fmt.Errorf("Blob is not defined")
	}
	if len(oc.rawBody) > 0 {
		return oc.rawBody, nil
	}
	return json.Marshal(oc.image)
}

// UnmarshalJSON extracts json content and populates the content.
func (oc *BOCIConfig) UnmarshalJSON(data []byte) error {
	image := v1.Image{}
	err := json.Unmarshal(data, &image)
	if err != nil {
		return err
	}
	oc.image = image
	oc.rawBody = make([]byte, len(data))
	copy(oc.rawBody, data)
	if oc.desc.MediaType == "" {
		oc.desc.MediaType = mediatype.OCI1ImageConfig
	}
	oc.desc.Digest = oc.desc.DigestAlgo().FromBytes(oc.rawBody)
	oc.desc.Size = int64(len(oc.rawBody))
	oc.blobSet = true
	return nil
}
