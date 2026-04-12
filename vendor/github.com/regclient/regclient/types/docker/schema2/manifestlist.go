package schema2

import (
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker"
	"github.com/regclient/regclient/types/mediatype"
)

// ManifestListSchemaVersion is a pre-configured versioned field for manifest lists
var ManifestListSchemaVersion = docker.Versioned{
	SchemaVersion: 2,
	MediaType:     mediatype.Docker2ManifestList,
}

// ManifestList references manifests for various platforms.
type ManifestList struct {
	docker.Versioned

	// Manifests lists descriptors in the manifest list
	Manifests []descriptor.Descriptor `json:"manifests"`

	// Annotations contains arbitrary metadata for the image index.
	// Note, this is not a defined docker schema2 field.
	Annotations map[string]string `json:"annotations,omitempty"`
}
