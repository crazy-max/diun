package reg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/internal/limitread"
	"github.com/regclient/regclient/internal/reghttp"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/warning"
)

// ManifestDelete removes a manifest by reference (digest) from a registry.
// This will implicitly delete all tags pointing to that manifest.
func (reg *Reg) ManifestDelete(ctx context.Context, r ref.Ref, opts ...scheme.ManifestOpts) error {
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	if r.Digest == "" {
		return fmt.Errorf("digest required to delete manifest, reference %s%.0w", r.CommonName(), errs.ErrMissingDigest)
	}

	mc := scheme.ManifestConfig{}
	for _, opt := range opts {
		opt(&mc)
	}

	if mc.CheckReferrers && mc.Manifest == nil {
		m, err := reg.ManifestGet(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to pull manifest for refers: %w", err)
		}
		mc.Manifest = m
	}
	if mc.Manifest != nil {
		if mr, ok := mc.Manifest.(manifest.Subjecter); ok {
			sDesc, err := mr.GetSubject()
			if err == nil && sDesc != nil && sDesc.Digest != "" {
				// attempt to delete the referrer, but ignore if the referrer entry wasn't found
				err = reg.referrerDelete(ctx, r, mc.Manifest)
				if err != nil && !errors.Is(err, errs.ErrNotFound) {
					return err
				}
			}
		}
	}
	rCache := r.SetDigest(r.Digest)
	reg.cacheMan.Delete(rCache)

	// build/send request
	req := &reghttp.Req{
		MetaKind:   reqmeta.Query,
		Host:       r.Registry,
		NoMirrors:  true,
		Method:     "DELETE",
		Repository: r.Repository,
		Path:       "manifests/" + r.Digest,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete manifest %s: %w", r.CommonName(), err)
	}
	defer resp.Close()
	if resp.HTTPResponse().StatusCode != 202 {
		return fmt.Errorf("failed to delete manifest %s: %w", r.CommonName(), reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	return nil
}

// ManifestGet retrieves a manifest from the registry
func (reg *Reg) ManifestGet(ctx context.Context, r ref.Ref) (manifest.Manifest, error) {
	var tagOrDigest string
	if r.Digest != "" {
		rCache := r.SetDigest(r.Digest)
		if m, err := reg.cacheMan.Get(rCache); err == nil {
			return m, nil
		}
		tagOrDigest = r.Digest
	} else if r.Tag != "" {
		tagOrDigest = r.Tag
	} else {
		return nil, fmt.Errorf("reference missing tag and digest: %s%.0w", r.CommonName(), errs.ErrMissingTagOrDigest)
	}

	// build/send request
	headers := http.Header{
		"Accept": []string{
			mediatype.OCI1ManifestList,
			mediatype.OCI1Manifest,
			mediatype.Docker2ManifestList,
			mediatype.Docker2Manifest,
			mediatype.Docker1ManifestSigned,
			mediatype.Docker1Manifest,
			mediatype.OCI1Artifact,
		},
	}
	req := &reghttp.Req{
		MetaKind:   reqmeta.Manifest,
		Host:       r.Registry,
		Method:     "GET",
		Repository: r.Repository,
		Path:       "manifests/" + tagOrDigest,
		Headers:    headers,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest %s: %w", r.CommonName(), err)
	}
	defer resp.Close()
	if resp.HTTPResponse().StatusCode != 200 {
		return nil, fmt.Errorf("failed to get manifest %s: %w", r.CommonName(), reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	// limit length
	size, _ := strconv.Atoi(resp.HTTPResponse().Header.Get("Content-Length"))
	if size > 0 && reg.manifestMaxPull > 0 && int64(size) > reg.manifestMaxPull {
		return nil, fmt.Errorf("manifest too large, received %d, limit %d: %s%.0w", size, reg.manifestMaxPull, r.CommonName(), errs.ErrSizeLimitExceeded)
	}
	rdr := &limitread.LimitRead{
		Reader: resp,
		Limit:  reg.manifestMaxPull,
	}

	// read manifest
	rawBody, err := io.ReadAll(rdr)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest for %s: %w", r.CommonName(), err)
	}

	m, err := manifest.New(
		manifest.WithRef(r),
		manifest.WithHeader(resp.HTTPResponse().Header),
		manifest.WithRaw(rawBody),
	)
	if err != nil {
		return nil, err
	}
	rCache := r.SetDigest(m.GetDescriptor().Digest.String())
	reg.cacheMan.Set(rCache, m)
	return m, nil
}

// ManifestHead returns metadata on the manifest from the registry
func (reg *Reg) ManifestHead(ctx context.Context, r ref.Ref) (manifest.Manifest, error) {
	// build the request
	var tagOrDigest string
	if r.Digest != "" {
		rCache := r.SetDigest(r.Digest)
		if m, err := reg.cacheMan.Get(rCache); err == nil {
			return m, nil
		}
		tagOrDigest = r.Digest
	} else if r.Tag != "" {
		tagOrDigest = r.Tag
	} else {
		return nil, fmt.Errorf("reference missing tag and digest: %s%.0w", r.CommonName(), errs.ErrMissingTagOrDigest)
	}

	// build/send request
	headers := http.Header{
		"Accept": []string{
			mediatype.OCI1ManifestList,
			mediatype.OCI1Manifest,
			mediatype.Docker2ManifestList,
			mediatype.Docker2Manifest,
			mediatype.Docker1ManifestSigned,
			mediatype.Docker1Manifest,
			mediatype.OCI1Artifact,
		},
	}
	req := &reghttp.Req{
		MetaKind:   reqmeta.Head,
		Host:       r.Registry,
		Method:     "HEAD",
		Repository: r.Repository,
		Path:       "manifests/" + tagOrDigest,
		Headers:    headers,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to request manifest head %s: %w", r.CommonName(), err)
	}
	defer resp.Close()
	if resp.HTTPResponse().StatusCode != 200 {
		return nil, fmt.Errorf("failed to request manifest head %s: %w", r.CommonName(), reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	return manifest.New(
		manifest.WithRef(r),
		manifest.WithHeader(resp.HTTPResponse().Header),
	)
}

// ManifestPut uploads a manifest to a registry
func (reg *Reg) ManifestPut(ctx context.Context, r ref.Ref, m manifest.Manifest, opts ...scheme.ManifestOpts) error {
	var tagOrDigest string
	if r.Digest != "" {
		tagOrDigest = r.Digest
	} else if r.Tag != "" {
		tagOrDigest = r.Tag
	} else {
		reg.slog.Warn("Manifest put requires a tag",
			slog.String("ref", r.Reference))
		return errs.ErrMissingTag
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}

	// create the request body
	mj, err := m.MarshalJSON()
	if err != nil {
		reg.slog.Warn("Error marshaling manifest",
			slog.String("ref", r.Reference),
			slog.String("err", err.Error()))
		return fmt.Errorf("error marshalling manifest for %s: %w", r.CommonName(), err)
	}

	// limit length
	if reg.manifestMaxPush > 0 && int64(len(mj)) > reg.manifestMaxPush {
		return fmt.Errorf("manifest too large, calculated %d, limit %d: %s%.0w", len(mj), reg.manifestMaxPush, r.CommonName(), errs.ErrSizeLimitExceeded)
	}

	// build/send request
	headers := http.Header{
		"Content-Type": []string{manifest.GetMediaType(m)},
	}
	q := url.Values{}
	if tagOrDigest == r.Tag && m.GetDescriptor().Digest.Algorithm() != digest.Canonical {
		// TODO(bmitch): EXPERIMENTAL parameter, registry support and OCI spec change needed
		q.Add(paramManifestDigest, m.GetDescriptor().Digest.String())
	}
	req := &reghttp.Req{
		MetaKind:   reqmeta.Manifest,
		Host:       r.Registry,
		NoMirrors:  true,
		Method:     "PUT",
		Repository: r.Repository,
		Path:       "manifests/" + tagOrDigest,
		Query:      q,
		Headers:    headers,
		BodyLen:    int64(len(mj)),
		BodyBytes:  mj,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to put manifest %s: %w", r.CommonName(), err)
	}
	err = resp.Close()
	if err != nil {
		return fmt.Errorf("failed to close request: %w", err)
	}
	if resp.HTTPResponse().StatusCode != 201 {
		return fmt.Errorf("failed to put manifest %s: %w", r.CommonName(), reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	rCache := r.SetDigest(m.GetDescriptor().Digest.String())
	reg.cacheMan.Set(rCache, m)

	// update referrers if defined on this manifest
	if mr, ok := m.(manifest.Subjecter); ok {
		mDesc, err := mr.GetSubject()
		if err != nil {
			return err
		}
		if mDesc != nil && mDesc.Digest.String() != "" {
			rSubj := r.SetDigest(mDesc.Digest.String())
			reg.cacheRL.Delete(rSubj)
			if mDesc.Digest.String() != resp.HTTPResponse().Header.Get(OCISubjectHeader) {
				err = reg.referrerPut(ctx, r, m)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
