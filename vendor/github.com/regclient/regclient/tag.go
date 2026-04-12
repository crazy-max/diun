package regclient

import (
	"context"
	"fmt"

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/tag"
)

// TagDelete deletes a tag from the registry. Since there's no API for this,
// you'd want to normally just delete the manifest. However multiple tags may
// point to the same manifest, so instead you must:
// 1. Make a manifest, for this we put a few labels and timestamps to be unique.
// 2. Push that manifest to the tag.
// 3. Delete the digest for that new manifest that is only used by that tag.
func (rc *RegClient) TagDelete(ctx context.Context, r ref.Ref) error {
	if !r.IsSet() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return err
	}
	return schemeAPI.TagDelete(ctx, r)
}

// TagList returns a tag list from a repository
func (rc *RegClient) TagList(ctx context.Context, r ref.Ref, opts ...scheme.TagOpts) (*tag.List, error) {
	if !r.IsSetRepo() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return nil, err
	}
	return schemeAPI.TagList(ctx, r, opts...)
}
