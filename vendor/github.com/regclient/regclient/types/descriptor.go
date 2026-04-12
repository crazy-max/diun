package types

import "github.com/regclient/regclient/types/descriptor"

type (
	// Descriptor is used in manifests to refer to content by media type, size, and digest.
	//
	// Deprecated: replace with [descriptor.Descriptor].
	//go:fix inline
	Descriptor = descriptor.Descriptor
	// MatchOpt defines conditions for a match descriptor.
	//
	// Deprecated: replace with [descriptor.MatchOpt].
	//go:fix inline
	MatchOpt = descriptor.MatchOpt
)

var (
	// EmptyData is the content of the empty JSON descriptor. See [mediatype.OCI1Empty].
	//
	// Deprecated: replace with [descriptor.EmptyData].
	//go:fix inline
	EmptyData = descriptor.EmptyData
	// EmptyDigest is the digest of the empty JSON descriptor. See [mediatype.OCI1Empty].
	//
	// Deprecated: replace with [descriptor.EmptyDigest].
	//go:fix inline
	EmptyDigest = descriptor.EmptyDigest
	// DescriptorListFilter returns a list of descriptors from the list matching the search options.
	// When opt.SortAnnotation is set, the order of descriptors with matching annotations is undefined.
	//
	// Deprecated: replace with [descriptor.DescriptorListFilter]
	//go:fix inline
	DescriptorListFilter = descriptor.DescriptorListFilter
	// DescriptorListSearch returns the first descriptor from the list matching the search options.
	//
	// Deprecated: replace with [descriptor.DescriptorListSearch]
	//go:fix inline
	DescriptorListSearch = descriptor.DescriptorListSearch
)
