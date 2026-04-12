package ocidir

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"slices"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"
)

// ManifestDelete removes a manifest, including all tags that point to that manifest
func (o *OCIDir) ManifestDelete(ctx context.Context, r ref.Ref, opts ...scheme.ManifestOpts) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if r.Digest == "" {
		return fmt.Errorf("digest required to delete manifest, reference %s%.0w", r.CommonName(), errs.ErrMissingDigest)
	}

	mc := scheme.ManifestConfig{}
	for _, opt := range opts {
		opt(&mc)
	}

	// always check for refers with ocidir
	if mc.Manifest == nil {
		m, err := o.manifestGet(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to pull manifest for refers: %w", err)
		}
		mc.Manifest = m
	}
	if mc.Manifest != nil {
		if ms, ok := mc.Manifest.(manifest.Subjecter); ok {
			sDesc, err := ms.GetSubject()
			if err == nil && sDesc != nil && sDesc.Digest != "" {
				// attempt to delete the referrer, but ignore if the referrer entry wasn't found
				err = o.referrerDelete(ctx, r, mc.Manifest)
				if err != nil && !errors.Is(err, errs.ErrNotFound) && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
		}
	}

	// get index
	changed := false
	index, err := o.readIndex(r, true)
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}
	for i := len(index.Manifests) - 1; i >= 0; i-- {
		// remove matching entry from index
		if r.Digest != "" && index.Manifests[i].Digest.String() == r.Digest {
			changed = true
			index.Manifests = slices.Delete(index.Manifests, i, i+1)
		}
	}
	// push manifest back out
	if changed {
		err = o.writeIndex(r, index, true)
		if err != nil {
			return fmt.Errorf("failed to write index: %w", err)
		}
	}

	// delete from filesystem like a registry would do
	d := digest.Digest(r.Digest)
	file := path.Join(r.Path, "blobs", d.Algorithm().String(), d.Encoded())
	err = os.Remove(file)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	o.refMod(r)
	return nil
}

// ManifestGet retrieves a manifest from a repository
func (o *OCIDir) ManifestGet(ctx context.Context, r ref.Ref) (manifest.Manifest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.manifestGet(ctx, r)
}

func (o *OCIDir) manifestGet(_ context.Context, r ref.Ref) (manifest.Manifest, error) {
	index, err := o.readIndex(r, true)
	if err != nil {
		return nil, fmt.Errorf("unable to read oci index: %w", err)
	}
	if r.Digest == "" && r.Tag == "" {
		r = r.SetTag("latest")
	}
	desc, err := indexGet(index, r)
	if err != nil {
		if r.Digest != "" {
			desc.Digest = digest.Digest(r.Digest)
		} else {
			return nil, err
		}
	}
	if desc.Digest == "" {
		return nil, errs.ErrNotFound
	}
	if err = desc.Digest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid digest in index: %s: %w", string(desc.Digest), err)
	}
	file := path.Join(r.Path, "blobs", desc.Digest.Algorithm().String(), desc.Digest.Encoded())
	//#nosec G304 users should validate references they attempt to open
	fd, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest: %w", err)
	}
	defer fd.Close()
	mb, err := io.ReadAll(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	if desc.Size == 0 {
		desc.Size = int64(len(mb))
	}
	o.slog.Debug("retrieved manifest",
		slog.String("ref", r.CommonName()),
		slog.String("file", file))
	return manifest.New(
		manifest.WithRef(r),
		manifest.WithDesc(desc),
		manifest.WithRaw(mb),
	)
}

// ManifestHead gets metadata about the manifest (existence, digest, mediatype, size)
func (o *OCIDir) ManifestHead(ctx context.Context, r ref.Ref) (manifest.Manifest, error) {
	index, err := o.readIndex(r, false)
	if err != nil {
		return nil, fmt.Errorf("unable to read oci index: %w", err)
	}
	if r.Digest == "" && r.Tag == "" {
		r = r.SetTag("latest")
	}
	desc, err := indexGet(index, r)
	if err != nil {
		if r.Digest != "" {
			desc.Digest = digest.Digest(r.Digest)
		} else {
			return nil, err
		}
	}
	if desc.Digest == "" {
		return nil, errs.ErrNotFound
	}
	if err = desc.Digest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid digest in index: %s: %w", string(desc.Digest), err)
	}
	// verify underlying file exists
	file := path.Join(r.Path, "blobs", desc.Digest.Algorithm().String(), desc.Digest.Encoded())
	fi, err := os.Stat(file)
	if err != nil || fi.IsDir() {
		return nil, errs.ErrNotFound
	}
	// if missing, set media type on desc
	if desc.MediaType == "" {
		//#nosec G304 users should validate references they attempt to open
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		mt := struct {
			MediaType     string `json:"mediaType,omitempty"`
			SchemaVersion int    `json:"schemaVersion,omitempty"`
			Signatures    []any  `json:"signatures,omitempty"`
		}{}
		err = json.Unmarshal(raw, &mt)
		if err != nil {
			return nil, err
		}
		if mt.MediaType != "" {
			desc.MediaType = mt.MediaType
			desc.Size = int64(len(raw))
		} else if mt.SchemaVersion == 1 && len(mt.Signatures) > 0 {
			desc.MediaType = mediatype.Docker1ManifestSigned
		} else if mt.SchemaVersion == 1 {
			desc.MediaType = mediatype.Docker1Manifest
			desc.Size = int64(len(raw))
		}
	}
	return manifest.New(
		manifest.WithRef(r),
		manifest.WithDesc(desc),
	)
}

// ManifestPut sends a manifest to the repository
func (o *OCIDir) ManifestPut(ctx context.Context, r ref.Ref, m manifest.Manifest, opts ...scheme.ManifestOpts) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.manifestPut(ctx, r, m, opts...)
}

func (o *OCIDir) manifestPut(ctx context.Context, r ref.Ref, m manifest.Manifest, opts ...scheme.ManifestOpts) error {
	config := scheme.ManifestConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	if !config.Child && r.Digest == "" && r.Tag == "" {
		r = r.SetTag("latest")
	}
	err := o.initIndex(r, true)
	if err != nil {
		return err
	}
	desc := m.GetDescriptor()
	if err = desc.Digest.Validate(); err != nil {
		return fmt.Errorf("invalid digest for manifest: %s: %w", string(desc.Digest), err)
	}
	b, err := m.RawBody()
	if err != nil {
		return fmt.Errorf("could not serialize manifest: %w", err)
	}
	if r.Digest != "" && desc.Digest.String() != r.Digest {
		// Digest algorithm may have changed, try recreating the manifest with the provided ref.
		// This will fail if the ref digest does not match the manifest.
		m, err = manifest.New(manifest.WithRef(r), manifest.WithRaw(b))
		if err != nil {
			return fmt.Errorf("failed to rebuilding manifest with ref \"%s\": %w", r.CommonName(), err)
		}
	}
	if r.Tag != "" {
		desc.Annotations = map[string]string{
			aOCIRefName: r.Tag,
		}
	}
	// create manifest CAS file
	dir := path.Join(r.Path, "blobs", desc.Digest.Algorithm().String())
	//#nosec G301 defer to user umask settings
	err = os.MkdirAll(dir, 0o777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("failed creating %s: %w", dir, err)
	}
	// write to a tmp file, rename after validating
	tmpFile, err := os.CreateTemp(dir, desc.Digest.Encoded()+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create manifest tmpfile: %w", err)
	}
	fi, err := tmpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat manifest tmpfile: %w", err)
	}
	tmpName := fi.Name()
	_, err = tmpFile.Write(b)
	errC := tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write manifest tmpfile: %w", err)
	}
	if errC != nil {
		return fmt.Errorf("failed to close manifest tmpfile: %w", errC)
	}
	file := path.Join(dir, desc.Digest.Encoded())
	//#nosec G703 inputs are user controlled
	err = os.Rename(path.Join(dir, tmpName), file)
	if err != nil {
		return fmt.Errorf("failed to write manifest (rename tmpfile): %w", err)
	}

	// verify/update index
	err = o.updateIndex(r, desc, config.Child, true)
	if err != nil {
		return err
	}
	o.refMod(r)
	o.slog.Debug("pushed manifest",
		slog.String("ref", r.CommonName()),
		slog.String("file", file))

	// update referrers if defined on this manifest
	if ms, ok := m.(manifest.Subjecter); ok {
		mDesc, err := ms.GetSubject()
		if err != nil {
			return err
		}
		if mDesc != nil && mDesc.Digest != "" {
			err = o.referrerPut(ctx, r, m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
