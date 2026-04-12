package ocidir

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"

	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
)

// BlobDelete removes a blob from the repository.
// This method does not verify that blobs are unused.
// Calling the [OCIDir.Close] method to trigger the garbage collection is preferred.
func (o *OCIDir) BlobDelete(ctx context.Context, r ref.Ref, d descriptor.Descriptor) error {
	err := d.Digest.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate digest %s: %w", d.Digest.String(), err)
	}
	file := path.Join(r.Path, "blobs", d.Digest.Algorithm().String(), d.Digest.Encoded())
	return os.Remove(file)
}

// BlobGet retrieves a blob, returning a reader
func (o *OCIDir) BlobGet(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error) {
	err := d.Digest.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate digest %s: %w", d.Digest.String(), err)
	}
	file := path.Join(r.Path, "blobs", d.Digest.Algorithm().String(), d.Digest.Encoded())
	//#nosec G304 users should validate references they attempt to open
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	if d.Size <= 0 {
		fi, err := fd.Stat()
		if err != nil {
			_ = fd.Close()
			return nil, err
		}
		d.Size = fi.Size()
	}
	br := blob.NewReader(
		blob.WithRef(r),
		blob.WithReader(fd),
		blob.WithDesc(d),
	)
	o.slog.Debug("retrieved blob",
		slog.String("ref", r.CommonName()),
		slog.String("file", file))
	return br, nil
}

// BlobHead verifies the existence of a blob, the reader contains the headers but no body to read
func (o *OCIDir) BlobHead(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error) {
	err := d.Digest.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate digest %s: %w", d.Digest.String(), err)
	}
	file := path.Join(r.Path, "blobs", d.Digest.Algorithm().String(), d.Digest.Encoded())
	//#nosec G304 users should validate references they attempt to open
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if d.Size <= 0 {
		fi, err := fd.Stat()
		if err != nil {
			return nil, err
		}
		d.Size = fi.Size()
	}
	br := blob.NewReader(
		blob.WithRef(r),
		blob.WithDesc(d),
	)
	return br, nil
}

// BlobMount attempts to perform a server side copy of the blob
func (o *OCIDir) BlobMount(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor) error {
	return errs.ErrUnsupported
}

// BlobPut sends a blob to the repository, returns the digest and size when successful
func (o *OCIDir) BlobPut(ctx context.Context, r ref.Ref, d descriptor.Descriptor, rdr io.Reader) (descriptor.Descriptor, error) {
	t := o.throttleGet(r, false)
	done, err := t.Acquire(ctx, reqmeta.Data{Kind: reqmeta.Blob, Size: d.Size})
	if err != nil {
		return d, err
	}
	defer done()

	err = o.initIndex(r, false)
	if err != nil {
		return d, err
	}
	digester := d.DigestAlgo().Digester()
	rdr = io.TeeReader(rdr, digester.Hash())
	// write the blob to a tmp file
	dir := path.Join(r.Path, "blobs", d.DigestAlgo().String())
	tmpPattern := "*.tmp"
	//#nosec G301 defer to user umask settings
	err = os.MkdirAll(dir, 0o777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return d, fmt.Errorf("failed creating %s: %w", dir, err)
	}
	tmpFile, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return d, fmt.Errorf("failed creating blob tmp file: %w", err)
	}
	fi, err := tmpFile.Stat()
	if err != nil {
		return d, fmt.Errorf("failed to stat blob tmpfile: %w", err)
	}
	tmpName := fi.Name()
	i, err := io.Copy(tmpFile, rdr)
	errC := tmpFile.Close()
	if err != nil {
		return d, err
	}
	if errC != nil {
		return d, errC
	}
	// validate result matches descriptor, or update descriptor if it wasn't defined
	if d.Digest.Validate() != nil {
		d.Digest = digester.Digest()
	} else if d.Digest != digester.Digest() {
		return d, fmt.Errorf("unexpected digest, expected %s, computed %s", d.Digest, digester.Digest())
	}
	if d.Size <= 0 {
		d.Size = i
	} else if i != d.Size {
		return d, fmt.Errorf("unexpected blob length, expected %d, received %d", d.Size, i)
	}
	file := path.Join(r.Path, "blobs", d.Digest.Algorithm().String(), d.Digest.Encoded())
	//#nosec G703 inputs are user controlled
	err = os.Rename(path.Join(dir, tmpName), file)
	if err != nil {
		return d, fmt.Errorf("failed to write blob (rename tmp file %s to %s): %w", path.Join(dir, tmpName), file, err)
	}
	o.slog.Debug("pushed blob",
		slog.String("ref", r.CommonName()),
		slog.String("file", file))

	o.mu.Lock()
	o.refMod(r)
	o.mu.Unlock()
	return d, nil
}
