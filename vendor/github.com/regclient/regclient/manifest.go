package regclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/warning"
)

type manifestOpt struct {
	d             descriptor.Descriptor
	platform      *platform.Platform
	schemeOpts    []scheme.ManifestOpts
	requireDigest bool
}

// ManifestOpts define options for the Manifest* commands.
type ManifestOpts func(*manifestOpt)

// WithManifest passes a manifest to ManifestDelete.
func WithManifest(m manifest.Manifest) ManifestOpts {
	return func(opts *manifestOpt) {
		opts.schemeOpts = append(opts.schemeOpts, scheme.WithManifest(m))
	}
}

// WithManifestCheckReferrers checks for referrers field on ManifestDelete.
// This will update the client managed referrer listing.
func WithManifestCheckReferrers() ManifestOpts {
	return func(opts *manifestOpt) {
		opts.schemeOpts = append(opts.schemeOpts, scheme.WithManifestCheckReferrers())
	}
}

// WithManifestChild for ManifestPut indicates the manifest is not the top level manifest being copied.
// This is used by the ocidir scheme to determine what entries to include in the index.json.
func WithManifestChild() ManifestOpts {
	return func(opts *manifestOpt) {
		opts.schemeOpts = append(opts.schemeOpts, scheme.WithManifestChild())
	}
}

// WithManifestDesc includes the descriptor for ManifestGet.
// This is used to automatically extract a Data field if available.
func WithManifestDesc(d descriptor.Descriptor) ManifestOpts {
	return func(opts *manifestOpt) {
		opts.d = d
	}
}

// WithManifestPlatform resolves the platform specific manifest on Get and Head requests.
// This causes an additional GET query to a registry when an Index or Manifest List is encountered.
// This option is ignored if the retrieved manifest is not an Index or Manifest List.
func WithManifestPlatform(p platform.Platform) ManifestOpts {
	return func(opts *manifestOpt) {
		opts.platform = &p
	}
}

// WithManifestRequireDigest falls back from a HEAD to a GET request when digest headers aren't received.
func WithManifestRequireDigest() ManifestOpts {
	return func(opts *manifestOpt) {
		opts.requireDigest = true
	}
}

// ManifestDelete removes a manifest, including all tags pointing to that registry.
// The reference must include the digest to delete (see TagDelete for deleting a tag).
// All tags pointing to the manifest will be deleted.
func (rc *RegClient) ManifestDelete(ctx context.Context, r ref.Ref, opts ...ManifestOpts) error {
	if !r.IsSet() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	opt := manifestOpt{schemeOpts: []scheme.ManifestOpts{}}
	for _, fn := range opts {
		fn(&opt)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return err
	}
	return schemeAPI.ManifestDelete(ctx, r, opt.schemeOpts...)
}

// ManifestGet retrieves a manifest.
func (rc *RegClient) ManifestGet(ctx context.Context, r ref.Ref, opts ...ManifestOpts) (manifest.Manifest, error) {
	if !r.IsSet() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	opt := manifestOpt{schemeOpts: []scheme.ManifestOpts{}}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.d.Digest != "" {
		r = r.AddDigest(opt.d.Digest.String())
		data, err := opt.d.GetData()
		if err == nil {
			return manifest.New(
				manifest.WithDesc(opt.d),
				manifest.WithRaw(data),
				manifest.WithRef(r),
			)
		}
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return nil, err
	}
	m, err := schemeAPI.ManifestGet(ctx, r)
	if err != nil {
		return m, err
	}
	if opt.platform != nil && !m.IsList() {
		rc.slog.Debug("ignoring platform option, image is not an index",
			slog.String("platform", opt.platform.String()),
			slog.String("ref", r.CommonName()))
	}
	// this will loop to handle a nested index
	for opt.platform != nil && m.IsList() {
		d, err := manifest.GetPlatformDesc(m, opt.platform)
		if err != nil {
			return m, err
		}
		r = r.SetDigest(d.Digest.String())
		m, err = schemeAPI.ManifestGet(ctx, r)
		if err != nil {
			return m, err
		}
	}
	return m, err
}

// ManifestHead queries for the existence of a manifest and returns metadata (digest, media-type, size).
func (rc *RegClient) ManifestHead(ctx context.Context, r ref.Ref, opts ...ManifestOpts) (manifest.Manifest, error) {
	if !r.IsSet() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	opt := manifestOpt{schemeOpts: []scheme.ManifestOpts{}}
	for _, fn := range opts {
		fn(&opt)
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return nil, err
	}
	m, err := schemeAPI.ManifestHead(ctx, r)
	if err != nil {
		return m, err
	}
	if opt.platform != nil && !m.IsList() {
		rc.slog.Debug("ignoring platform option, image is not an index",
			slog.String("platform", opt.platform.String()),
			slog.String("ref", r.CommonName()))
	}
	// this will loop to handle a nested index
	for opt.platform != nil && m.IsList() {
		if !m.IsSet() {
			m, err = schemeAPI.ManifestGet(ctx, r)
		}
		d, err := manifest.GetPlatformDesc(m, opt.platform)
		if err != nil {
			return m, err
		}
		r = r.SetDigest(d.Digest.String())
		m, err = schemeAPI.ManifestHead(ctx, r)
		if err != nil {
			return m, err
		}
	}
	if opt.requireDigest && m.GetDescriptor().Digest.String() == "" {
		m, err = schemeAPI.ManifestGet(ctx, r)
	}
	return m, err
}

// ManifestPut pushes a manifest.
// Any descriptors referenced by the manifest typically need to be pushed first.
func (rc *RegClient) ManifestPut(ctx context.Context, r ref.Ref, m manifest.Manifest, opts ...ManifestOpts) error {
	if !r.IsSetRepo() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	opt := manifestOpt{schemeOpts: []scheme.ManifestOpts{}}
	for _, fn := range opts {
		fn(&opt)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return err
	}
	return schemeAPI.ManifestPut(ctx, r, m, opt.schemeOpts...)
}
