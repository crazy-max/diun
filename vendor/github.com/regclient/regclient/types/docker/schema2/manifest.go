package schema2

import (
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker"
	"github.com/regclient/regclient/types/mediatype"
)

// ManifestSchemaVersion is a pre-configured versioned field for manifests
var ManifestSchemaVersion = docker.Versioned{
	SchemaVersion: 2,
	MediaType:     mediatype.Docker2Manifest,
}

// Manifest defines a schema2 manifest.
type Manifest struct {
	docker.Versioned

	// Config references the image configuration as a blob.
	Config descriptor.Descriptor `json:"config"`

	// Layers lists descriptors for the layers referenced by the
	// configuration.
	Layers []descriptor.Descriptor `json:"layers"`

	// Annotations contains arbitrary metadata for the image index.
	// Note, this is not a defined docker schema2 field.
	Annotations map[string]string `json:"annotations,omitempty"`
}
