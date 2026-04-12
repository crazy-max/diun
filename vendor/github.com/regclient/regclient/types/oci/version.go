// Package oci defines common settings for all OCI types
package oci

// Versioned provides a struct with the manifest schemaVersion and mediaType.
type Versioned struct {
	// SchemaVersion is the image manifest schema that this image follows
	SchemaVersion int `json:"schemaVersion"`
}
