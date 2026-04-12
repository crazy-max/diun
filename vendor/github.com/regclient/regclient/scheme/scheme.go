// Package scheme defines the interface for various reference schemes.
package scheme

import (
	"context"
	"io"

	"github.com/regclient/regclient/internal/pqueue"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/referrer"
	"github.com/regclient/regclient/types/tag"
)

// API is used to interface between different methods to store images.
type API interface {
	// BlobDelete removes a blob from the repository.
	BlobDelete(ctx context.Context, r ref.Ref, d descriptor.Descriptor) error
	// BlobGet retrieves a blob, returning a reader.
	BlobGet(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error)
	// BlobHead verifies the existence of a blob, the reader contains the headers but no body to read.
	BlobHead(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error)
	// BlobMount attempts to perform a server side copy of the blob.
	BlobMount(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor) error
	// BlobPut sends a blob to the repository, returns the digest and size when successful.
	BlobPut(ctx context.Context, r ref.Ref, d descriptor.Descriptor, rdr io.Reader) (descriptor.Descriptor, error)

	// ManifestDelete removes a manifest, including all tags that point to that manifest.
	ManifestDelete(ctx context.Context, r ref.Ref, opts ...ManifestOpts) error
	// ManifestGet retrieves a manifest from a repository.
	ManifestGet(ctx context.Context, r ref.Ref) (manifest.Manifest, error)
	// ManifestHead gets metadata about the manifest (existence, digest, mediatype, size).
	ManifestHead(ctx context.Context, r ref.Ref) (manifest.Manifest, error)
	// ManifestPut sends a manifest to the repository.
	ManifestPut(ctx context.Context, r ref.Ref, m manifest.Manifest, opts ...ManifestOpts) error

	// Ping verifies access to a registry or equivalent.
	Ping(ctx context.Context, r ref.Ref) (ping.Result, error)

	// ReferrerList returns a list of referrers to a given reference.
	ReferrerList(ctx context.Context, r ref.Ref, opts ...ReferrerOpts) (referrer.ReferrerList, error)

	// TagDelete removes a tag from the repository.
	TagDelete(ctx context.Context, r ref.Ref) error
	// TagList returns a list of tags from the repository.
	TagList(ctx context.Context, r ref.Ref, opts ...TagOpts) (*tag.List, error)
}

// Closer is used to check if a scheme implements the Close API.
type Closer interface {
	Close(ctx context.Context, r ref.Ref) error
}

// GCLocker is used to indicate locking is available for GC management.
type GCLocker interface {
	// GCLock a reference to prevent GC from triggering during a put, locks are not exclusive.
	GCLock(r ref.Ref)
	// GCUnlock a reference to allow GC (once all locks are released).
	// The reference should be closed after this step and unlock should only be called once per each Lock call.
	GCUnlock(r ref.Ref)
}

// Throttler is used to indicate the scheme implements Throttle.
type Throttler interface {
	Throttle(r ref.Ref, put bool) []*pqueue.Queue[reqmeta.Data]
}

// ManifestConfig is used by schemes to import [ManifestOpts].
type ManifestConfig struct {
	CheckReferrers bool
	Child          bool // used when pushing a child of a manifest list, skips indexing in ocidir
	Manifest       manifest.Manifest
}

// ManifestOpts is used to set options on manifest APIs.
type ManifestOpts func(*ManifestConfig)

// WithManifestCheckReferrers is used when deleting a manifest.
// It indicates the manifest should be fetched and referrers should be deleted if defined.
func WithManifestCheckReferrers() ManifestOpts {
	return func(config *ManifestConfig) {
		config.CheckReferrers = true
	}
}

// WithManifestChild indicates the API call is on a child manifest.
// This is used internally when copying multi-platform manifests.
// This bypasses tracking of an untagged digest in ocidir which is needed for garbage collection.
func WithManifestChild() ManifestOpts {
	return func(config *ManifestConfig) {
		config.Child = true
	}
}

// WithManifest is used to pass the manifest to a method to avoid an extra GET request.
// This is used on a delete to check for referrers.
func WithManifest(m manifest.Manifest) ManifestOpts {
	return func(mc *ManifestConfig) {
		mc.Manifest = m
	}
}

// ReferrerConfig is used by schemes to import [ReferrerOpts].
type ReferrerConfig struct {
	MatchOpt descriptor.MatchOpt // filter/sort results
	Platform string              // get referrers for a specific platform
	SrcRepo  ref.Ref             // repo used to query referrers
}

// ReferrerOpts is used to set options on referrer APIs.
type ReferrerOpts func(*ReferrerConfig)

// WithReferrerMatchOpt filters results using [descriptor.MatchOpt].
func WithReferrerMatchOpt(mo descriptor.MatchOpt) ReferrerOpts {
	return func(config *ReferrerConfig) {
		config.MatchOpt = config.MatchOpt.Merge(mo)
	}
}

// WithReferrerPlatform gets referrers for a single platform from a multi-platform manifest.
// Note that this is implemented by [regclient.ReferrerList] and not the individual scheme implementations.
func WithReferrerPlatform(p string) ReferrerOpts {
	return func(config *ReferrerConfig) {
		config.Platform = p
	}
}

// WithReferrerSource pulls referrers from a separate source.
// Note that this is implemented by [regclient.ReferrerList] and not the individual scheme implementations.
func WithReferrerSource(r ref.Ref) ReferrerOpts {
	return func(config *ReferrerConfig) {
		config.SrcRepo = r
	}
}

// WithReferrerAT filters by a specific artifactType value.
//
// Deprecated: replace with [WithReferrerMatchOpt].
//
//go:fix inline
func WithReferrerAT(at string) ReferrerOpts {
	return WithReferrerMatchOpt(descriptor.MatchOpt{ArtifactType: at})
}

// WithReferrerAnnotations filters by a list of annotations, all of which must match.
//
// Deprecated: replace with [WithReferrerMatchOpt].
//
//go:fix inline
func WithReferrerAnnotations(annotations map[string]string) ReferrerOpts {
	return WithReferrerMatchOpt(descriptor.MatchOpt{Annotations: annotations})
}

// WithReferrerSort orders the resulting referrers listing according to a specified annotation.
//
// Deprecated: replace with [WithReferrerMatchOpt].
//
//go:fix inline
func WithReferrerSort(annotation string, desc bool) ReferrerOpts {
	return WithReferrerMatchOpt(descriptor.MatchOpt{SortAnnotation: annotation, SortDesc: desc})
}

// ReferrerFilter filters the referrer list according to the config.
func ReferrerFilter(config ReferrerConfig, rlIn referrer.ReferrerList) referrer.ReferrerList {
	return referrer.ReferrerList{
		Subject:     rlIn.Subject,
		Source:      rlIn.Source,
		Manifest:    rlIn.Manifest,
		Annotations: rlIn.Annotations,
		Tags:        rlIn.Tags,
		Descriptors: descriptor.DescriptorListFilter(rlIn.Descriptors, config.MatchOpt),
	}
}

// RepoConfig is used by schemes to import [RepoOpts].
type RepoConfig struct {
	Limit int
	Last  string
}

// RepoOpts is used to set options on repo APIs.
type RepoOpts func(*RepoConfig)

// WithRepoLimit passes a maximum number of repositories to return to the repository list API.
// Registries may ignore this.
func WithRepoLimit(l int) RepoOpts {
	return func(config *RepoConfig) {
		config.Limit = l
	}
}

// WithRepoLast passes the last received repository for requesting the next batch of repositories.
// Registries may ignore this.
func WithRepoLast(l string) RepoOpts {
	return func(config *RepoConfig) {
		config.Last = l
	}
}

// TagConfig is used by schemes to import [TagOpts].
type TagConfig struct {
	Limit int
	Last  string
}

// TagOpts is used to set options on tag APIs.
type TagOpts func(*TagConfig)

// WithTagLimit passes a maximum number of tags to return to the tag list API.
// Registries may ignore this.
func WithTagLimit(limit int) TagOpts {
	return func(t *TagConfig) {
		t.Limit = limit
	}
}

// WithTagLast passes the last received tag for requesting the next batch of tags.
// Registries may ignore this.
func WithTagLast(last string) TagOpts {
	return func(t *TagConfig) {
		t.Last = last
	}
}
