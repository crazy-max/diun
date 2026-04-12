package reg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/regclient/regclient/internal/httplink"
	"github.com/regclient/regclient/internal/reghttp"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/referrer"
	"github.com/regclient/regclient/types/warning"
)

const OCISubjectHeader = "OCI-Subject"

// ReferrerList returns a list of referrers to a given reference.
// The reference must include the digest. Use [regclient.ReferrerList] to resolve the platform or tag.
func (reg *Reg) ReferrerList(ctx context.Context, rSubject ref.Ref, opts ...scheme.ReferrerOpts) (referrer.ReferrerList, error) {
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
	rl := referrer.ReferrerList{
		Tags: []string{},
	}
	if rSubject.Digest == "" {
		return rl, fmt.Errorf("digest required to query referrers %s", rSubject.CommonName())
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}

	found := false
	// try cache
	rl, err := reg.cacheRL.Get(r)
	if err == nil {
		found = true
	}
	// try referrers API
	if !found {
		referrerEnabled, ok := reg.featureGet("referrer", r.Registry, r.Repository)
		if !ok || referrerEnabled {
			// attempt to call the referrer API
			rl, err = reg.referrerListByAPI(ctx, r, config)
			if !ok {
				// save the referrer API state
				reg.featureSet("referrer", r.Registry, r.Repository, err == nil)
			}
			if err == nil {
				if config.MatchOpt.ArtifactType == "" {
					// only cache if successful and artifactType is not filtered
					reg.cacheRL.Set(r, rl)
				}
				found = true
			}
		}
	}
	// fall back to tag
	if !found {
		rl, err = reg.referrerListByTag(ctx, r)
		if err == nil {
			reg.cacheRL.Set(r, rl)
		}
	}
	rl.Subject = rSubject
	if config.SrcRepo.IsSet() {
		rl.Source = config.SrcRepo
	}
	if err != nil {
		return rl, err
	}

	// apply client side filters and return result
	rl = scheme.ReferrerFilter(config, rl)
	return rl, nil
}

func (reg *Reg) referrerListByAPI(ctx context.Context, r ref.Ref, config scheme.ReferrerConfig) (referrer.ReferrerList, error) {
	rl := referrer.ReferrerList{
		Subject: r,
		Tags:    []string{},
	}
	var link *url.URL
	// loop for paging
	for {
		rlAdd, linkNext, err := reg.referrerListByAPIPage(ctx, r, config, link)
		if err != nil {
			return rl, err
		}
		if rl.Manifest == nil {
			rl = rlAdd
		} else {
			rl.Descriptors = append(rl.Descriptors, rlAdd.Descriptors...)
		}
		if linkNext == nil {
			break
		}
		link = linkNext
	}
	return rl, nil
}

func (reg *Reg) referrerListByAPIPage(ctx context.Context, r ref.Ref, config scheme.ReferrerConfig, link *url.URL) (referrer.ReferrerList, *url.URL, error) {
	rl := referrer.ReferrerList{
		Subject: r,
		Tags:    []string{},
	}
	query := url.Values{}
	if config.MatchOpt.ArtifactType != "" {
		query.Set("artifactType", config.MatchOpt.ArtifactType)
	}
	req := &reghttp.Req{
		MetaKind:   reqmeta.Query,
		Host:       r.Registry,
		Method:     "GET",
		Repository: r.Repository,
	}
	if link == nil {
		req.Path = "referrers/" + r.Digest
		req.Query = query
		req.IgnoreErr = true
	}
	if link != nil {
		req.DirectURL = link
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return rl, nil, fmt.Errorf("failed to get referrers %s: %w", r.CommonName(), err)
	}
	defer resp.Close()
	if resp.HTTPResponse().StatusCode != 200 {
		return rl, nil, fmt.Errorf("failed to get referrers %s: %w", r.CommonName(), reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	// read manifest
	rawBody, err := io.ReadAll(resp)
	if err != nil {
		return rl, nil, fmt.Errorf("error reading referrers for %s: %w", r.CommonName(), err)
	}

	m, err := manifest.New(
		manifest.WithRef(r.SetDigest("")),
		manifest.WithHeader(resp.HTTPResponse().Header),
		manifest.WithRaw(rawBody),
	)
	if err != nil {
		return rl, nil, err
	}
	ociML, ok := m.GetOrig().(v1.Index)
	if !ok {
		return rl, nil, fmt.Errorf("unexpected manifest type for referrers: %s, %w", m.GetDescriptor().MediaType, errs.ErrUnsupportedMediaType)
	}
	rl.Manifest = m
	rl.Descriptors = ociML.Manifests
	rl.Annotations = ociML.Annotations

	// lookup next link
	respHead := resp.HTTPResponse().Header
	links, err := httplink.Parse((respHead.Values("Link")))
	if err != nil {
		return rl, nil, err
	}
	next, err := links.Get("rel", "next")
	if err != nil {
		// no next link
		link = nil
	} else {
		link = resp.HTTPResponse().Request.URL
		if link == nil {
			return rl, nil, fmt.Errorf("referrers list failed to get URL of previous request")
		}
		link, err = link.Parse(next.URI)
		if err != nil {
			return rl, nil, fmt.Errorf("referrers list failed to parse Link: %w", err)
		}
	}

	return rl, link, nil
}

func (reg *Reg) referrerListByTag(ctx context.Context, r ref.Ref) (referrer.ReferrerList, error) {
	rl := referrer.ReferrerList{
		Subject: r,
		Tags:    []string{},
	}
	rlTag, err := referrer.FallbackTag(r)
	if err != nil {
		return rl, err
	}
	m, err := reg.ManifestGet(ctx, rlTag)
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
	// return resulting index
	rl.Manifest = m
	rl.Descriptors = ociML.Manifests
	rl.Annotations = ociML.Annotations
	rl.Tags = append(rl.Tags, rlTag.Tag)
	return rl, nil
}

// referrerDelete deletes a referrer associated with a manifest
func (reg *Reg) referrerDelete(ctx context.Context, r ref.Ref, m manifest.Manifest) error {
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// get subject field
	mSubject, ok := m.(manifest.Subjecter)
	if !ok {
		return fmt.Errorf("manifest does not support the subject field: %w", errs.ErrUnsupportedMediaType)
	}
	subject, err := mSubject.GetSubject()
	if err != nil {
		return err
	}
	// validate/set subject descriptor
	if subject == nil || subject.Digest == "" {
		return fmt.Errorf("refers is not set%.0w", errs.ErrNotFound)
	}

	// remove from cache
	rSubject := r.SetDigest(subject.Digest.String())
	reg.cacheRL.Delete(rSubject)

	// if referrer API is available, nothing to do, return
	if reg.referrerPing(ctx, rSubject) {
		return nil
	}

	// fallback to using tag schema for refers
	rl, err := reg.referrerListByTag(ctx, rSubject)
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
		err = reg.TagDelete(ctx, rlTag)
		if err == nil {
			return nil
		}
		// if delete is not supported, fall back to pushing empty list
	}
	return reg.ManifestPut(ctx, rlTag, rl.Manifest)
}

// referrerPut pushes a new referrer associated with a manifest
func (reg *Reg) referrerPut(ctx context.Context, r ref.Ref, m manifest.Manifest) error {
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// get subject field
	mSubject, ok := m.(manifest.Subjecter)
	if !ok {
		return fmt.Errorf("manifest does not support the subject field: %w", errs.ErrUnsupportedMediaType)
	}
	subject, err := mSubject.GetSubject()
	if err != nil {
		return err
	}
	// validate/set subject descriptor
	if subject == nil || subject.Digest == "" {
		return fmt.Errorf("subject is not set%.0w", errs.ErrNotFound)
	}

	// lock to avoid internal race conditions between pulling and pushing tag
	reg.muRefTag.Lock()
	defer reg.muRefTag.Unlock()
	// fallback to using tag schema for refers
	rSubject := r.SetDigest(subject.Digest.String())
	rl, err := reg.referrerListByTag(ctx, rSubject)
	if err != nil {
		return err
	}
	err = rl.Add(m)
	if err != nil {
		return err
	}
	// ensure the referrer list does not have a subject itself (avoiding circular locks)
	if ms, ok := rl.Manifest.(manifest.Subjecter); ok {
		mDesc, err := ms.GetSubject()
		if err != nil {
			return err
		}
		if mDesc != nil && mDesc.Digest != "" {
			return fmt.Errorf("fallback referrers manifest should not have a subject: %s", rSubject.CommonName())
		}
	}
	// push updated referrer list by tag
	rlTag, err := referrer.FallbackTag(rSubject)
	if err != nil {
		return err
	}
	if len(rl.Tags) == 0 {
		rl.Tags = []string{rlTag.Tag}
	}
	err = reg.ManifestPut(ctx, rlTag, rl.Manifest)
	if err == nil {
		reg.cacheRL.Set(rSubject, rl)
	}
	return err
}

// referrerPing verifies the registry supports the referrers API
func (reg *Reg) referrerPing(ctx context.Context, r ref.Ref) bool {
	referrerEnabled, ok := reg.featureGet("referrer", r.Registry, r.Repository)
	if ok {
		return referrerEnabled
	}
	req := &reghttp.Req{
		MetaKind:   reqmeta.Query,
		Host:       r.Registry,
		Method:     "GET",
		Repository: r.Repository,
		Path:       "referrers/" + r.Digest,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		reg.featureSet("referrer", r.Registry, r.Repository, false)
		return false
	}
	_ = resp.Close()
	result := resp.HTTPResponse().StatusCode == 200
	reg.featureSet("referrer", r.Registry, r.Repository, result)
	return result
}
