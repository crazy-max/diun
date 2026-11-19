// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ocidir

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/referrer"
)

// ReferrerList returns a list of referrers to a given reference.
// The reference must include the digest. Use [regclient.ReferrerList] to resolve the platform or tag.
func (o *OCIDir) ReferrerList(ctx context.Context, r ref.Ref, opts ...scheme.ReferrerOpts) (referrer.ReferrerList, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.referrerList(ctx, r, opts...)
}

func (o *OCIDir) referrerList(ctx context.Context, rSubject ref.Ref, opts ...scheme.ReferrerOpts) (referrer.ReferrerList, error) {
	config := scheme.ReferrerConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	var r ref.Ref
	if config.SrcRepo.IsSet() {
		r = config.SrcRepo.SetDigest(rSubject.Digest)
	} else {
		r = rSubject.SetDigest(rSubject.Digest)
	}
	rl := referrer.New(rSubject)
	if rSubject.Digest == "" {
		return rl, fmt.Errorf("digest required to query referrers %s", rSubject.CommonName())
	}

	// TODO: Add fast path with private field in index.json indicating referrers are tracked using an annotation.
	// This can be implemented with backward compatible support by also pushing individual manifests or the fallback tag.

	// TODO: Upstream OCI specification of how to track subject/referrers in an OCI Layout is still pending.

	// optional slow query of all manifests for ORAS support
	if config.SlowSearch {
		idx, err := o.readIndex(r, true)
		if err != nil {
			return rl, err
		}
		for _, desc := range idx.Manifests {
			// pull the manifest
			rCur := r.SetDigest(desc.Digest.String())
			m, err := o.manifestGet(ctx, rCur)
			if err != nil {
				// skip all errors
				continue
			}
			// check for the subject
			mSubj, ok := m.(manifest.Subjecter)
			if !ok {
				continue
			}
			subj, err := mSubj.GetSubject()
			if err != nil || subj == nil || subj.Digest.String() != rSubject.Digest {
				continue
			}
			// extract/track tag
			if name, ok := desc.Annotations[types.AnnotationRefName]; ok && !strings.ContainsAny(name, ":/") {
				rl.Tags = append(rl.Tags, name)
			}
			// if matching, add to rl
			err = rl.Add(m)
			if err != nil {
				return rl, err
			}
		}
	}

	// try to request the fallback tag
	rlFallback, err := o.referrerListFallback(ctx, r)
	if err == nil {
		err = rl.Merge(rlFallback)
		if err != nil {
			return rl, fmt.Errorf("failed to merge referrers from fallback tag: %w", err)
		}
	}

	// update referrer list
	rl.Subject = rSubject
	if config.SrcRepo.IsSet() {
		rl.Source = config.SrcRepo
	}
	rl = scheme.ReferrerFilter(config, rl)

	return rl, nil
}

func (o *OCIDir) referrerListFallback(ctx context.Context, r ref.Ref) (referrer.ReferrerList, error) {
	rl := referrer.New(r)
	rlTag, err := referrer.FallbackTag(r)
	if err != nil {
		return rl, err
	}
	m, err := o.manifestGet(ctx, rlTag)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			// empty list, initialize a new manifest
			rl.Manifest, err = manifest.New(manifest.WithOrig(v1.Index{
				Versioned: v1.IndexSchemaVersion,
				MediaType: mediatype.OCI1ManifestList,
			}))
			if err != nil {
				return rl, err
			}
			return rl, nil
		}
		return rl, err
	}
	ociML, ok := m.GetOrig().(v1.Index)
	if !ok {
		return rl, fmt.Errorf("manifest is not an OCI index: %s", rlTag.CommonName())
	}
	// update referrer list
	rl.Manifest = m
	rl.Descriptors = ociML.Manifests
	rl.Annotations = ociML.Annotations
	rl.Tags = append(rl.Tags, rlTag.Tag)
	return rl, nil
}

// referrerDelete deletes a referrer associated with a manifest
func (o *OCIDir) referrerDelete(ctx context.Context, r ref.Ref, m manifest.Manifest) error {
	// get refers field
	mSubject, ok := m.(manifest.Subjecter)
	if !ok {
		return fmt.Errorf("manifest does not support subject: %w", errs.ErrUnsupportedMediaType)
	}
	subject, err := mSubject.GetSubject()
	if err != nil {
		return err
	}
	// validate/set subject descriptor
	if subject == nil || subject.Digest == "" {
		return fmt.Errorf("subject is not set%.0w", errs.ErrNotFound)
	}

	// get descriptor for subject
	rSubject := r.SetDigest(subject.Digest.String())

	// pull existing referrer list
	rl, err := o.referrerList(ctx, rSubject)
	if err != nil {
		return err
	}
	err = rl.Delete(m)
	if err != nil {
		return err
	}

	// push updated referrer list by tag
	rlTag, err := referrer.FallbackTag(rSubject)
	if err != nil {
		return err
	}
	if rl.IsEmpty() {
		err = o.tagDelete(ctx, rlTag)
		if err == nil {
			return nil
		}
		// if delete is not supported, fall back to pushing empty list
	}
	return o.manifestPut(ctx, rlTag, rl.Manifest)
}

// referrerPut pushes a new referrer associated with a given reference
func (o *OCIDir) referrerPut(ctx context.Context, r ref.Ref, m manifest.Manifest) error {
	// get subject field
	mSubject, ok := m.(manifest.Subjecter)
	if !ok {
		return fmt.Errorf("manifest does not support subject: %w", errs.ErrUnsupportedMediaType)
	}
	subject, err := mSubject.GetSubject()
	if err != nil {
		return err
	}
	// validate/set subject descriptor
	if subject == nil || subject.Digest == "" {
		return fmt.Errorf("subject is not set%.0w", errs.ErrNotFound)
	}

	// get descriptor for subject
	rSubject := r.SetDigest(subject.Digest.String())

	// pull existing referrer list
	rl, err := o.referrerList(ctx, rSubject)
	if err != nil {
		return err
	}
	err = rl.Add(m)
	if err != nil {
		return err
	}

	// TODO: add support for pushing a fast path generated response

	// TODO: allow fallback tag creation to be skipped, potentially via a config in the manifestPut

	// push updated referrer list by tag
	rlTag, err := referrer.FallbackTag(rSubject)
	if err != nil {
		return err
	}
	return o.manifestPut(ctx, rlTag, rl.Manifest)
}
