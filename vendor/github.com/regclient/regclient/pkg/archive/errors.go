package archive

import "errors"

var (
	// ErrNotImplemented used for routines that need to be developed still
	ErrNotImplemented = errors.New("this archive routine is not implemented yet")
	// ErrUnknownType used for unknown compression types
	ErrUnknownType = errors.New("unknown compression type")
	// ErrXzUnsupported because there isn't a Go package for this and I'm
	// avoiding dependencies on external binaries
	ErrXzUnsupported = errors.New("xz compression is currently unsupported")
)
