package v1

import "github.com/regclient/regclient/types/descriptor"

// ArtifactManifest EXPERIMENTAL defines an OCI Artifact
type ArtifactManifest struct {
	// MediaType is the media type of the object this schema refers to.
	MediaType string `json:"mediaType"`

	// ArtifactType is the media type of the artifact this schema refers to.
	ArtifactType string `json:"artifactType,omitempty"`

	// Blobs is a collection of blobs referenced by this manifest.
	Blobs []descriptor.Descriptor `json:"blobs,omitempty"`

	// Subject is an optional link from the image manifest to another manifest forming an association between the image manifest and the other manifest.
	Subject *descriptor.Descriptor `json:"subject,omitempty"`

	// Annotations contains arbitrary metadata for the artifact manifest.
	Annotations map[string]string `json:"annotations,omitempty"`
}
