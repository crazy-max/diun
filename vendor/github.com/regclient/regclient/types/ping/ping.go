// Package ping is used for data types with the Ping methods.
package ping

import (
	"io/fs"
	"net/http"
)

// Result is the response to a ping request.
type Result struct {
	Header http.Header // Header is defined for responses from a registry.
	Stat   fs.FileInfo // Stat is defined for responses from an ocidir.
}
