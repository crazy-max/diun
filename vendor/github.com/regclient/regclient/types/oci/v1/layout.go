package v1

// ImageLayout is the structure in the "oci-layout" file, found in the root of an OCI Image-layout directory.
type ImageLayout struct {
	Version string `json:"imageLayoutVersion"`
}
