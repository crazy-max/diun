package v1

import (
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/oci"
)

// ManifestSchemaVersion is a pre-configured versioned field for manifests
var ManifestSchemaVersion = oci.Versioned{
	SchemaVersion: 2,
}

// Manifest defines an OCI image
type Manifest struct {
	oci.Versioned

	// MediaType specifies the type of this document data structure e.g. `application/vnd.oci.image.manifest.v1+json`
	MediaType string `json:"mediaType,omitempty"`

	// ArtifactType specifies the IANA media type of artifact when the manifest is used for an artifact.
	ArtifactType string `json:"artifactType,omitempty"`

	// Config references a configuration object for a container, by digest.
	// The referenced configuration object is a JSON blob that the runtime uses to set up the container.
	Config descriptor.Descriptor `json:"config"`

	// Layers is an indexed list of layers referenced by the manifest.
	Layers []descriptor.Descriptor `json:"layers"`

	// Subject is an optional link from the image manifest to another manifest forming an association between the image manifest and the other manifest.
	Subject *descriptor.Descriptor `json:"subject,omitempty"`

	// Annotations contains arbitrary metadata for the image manifest.
	Annotations map[string]string `json:"annotations,omitempty"`
}
