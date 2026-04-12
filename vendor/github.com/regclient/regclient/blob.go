package regclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/regclient/regclient/internal/pqueue"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/warning"
)

const blobCBFreq = time.Millisecond * 100

type blobOpt struct {
	callback   func(kind types.CallbackKind, instance string, state types.CallbackState, cur, total int64)
	readerHook func(*blob.BReader) (*blob.BReader, error)
}

// BlobOpts define options for the Image* commands.
type BlobOpts func(*blobOpt)

// BlobWithCallback provides progress data to a callback function.
func BlobWithCallback(callback func(kind types.CallbackKind, instance string, state types.CallbackState, cur, total int64)) BlobOpts {
	return func(opts *blobOpt) {
		opts.callback = callback
	}
}

// BlobWithReaderHook is called in [RegClient.BlobCopy] with the blob source.
// The returned [blob.BReader] is pushed to the target.
// If the hook returns an error, the copy will fail.
func BlobWithReaderHook(hook func(*blob.BReader) (*blob.BReader, error)) BlobOpts {
	return func(opts *blobOpt) {
		opts.readerHook = hook
	}
}

// BlobCopy copies a blob between two locations.
// If the blob already exists in the target, the copy is skipped.
// A server side cross repository blob mount is attempted.
func (rc *RegClient) BlobCopy(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor, opts ...BlobOpts) error {
	if !refSrc.IsSetRepo() {
		return fmt.Errorf("refSrc is not set: %s%.0w", refSrc.CommonName(), errs.ErrInvalidReference)
	}
	if !refTgt.IsSetRepo() {
		return fmt.Errorf("refTgt is not set: %s%.0w", refTgt.CommonName(), errs.ErrInvalidReference)
	}
	var opt blobOpt
	for _, optFn := range opts {
		optFn(&opt)
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	tDesc := d
	tDesc.URLs = []string{} // ignore URLs when pushing to target
	if opt.callback != nil {
		opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackStarted, 0, d.Size)
	}
	// for the same repository, there's nothing to copy
	if ref.EqualRepository(refSrc, refTgt) {
		if opt.callback != nil {
			opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackSkipped, 0, d.Size)
		}
		rc.slog.Debug("Blob copy skipped, same repo",
			slog.String("src", refSrc.Reference),
			slog.String("tgt", refTgt.Reference),
			slog.String("digest", string(d.Digest)))
		return nil
	}
	// check if layer already exists
	if _, err := rc.BlobHead(ctx, refTgt, tDesc); err == nil {
		if opt.callback != nil {
			opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackSkipped, 0, d.Size)
		}
		rc.slog.Debug("Blob copy skipped, already exists",
			slog.String("src", refSrc.Reference),
			slog.String("tgt", refTgt.Reference),
			slog.String("digest", string(d.Digest)))
		return nil
	}
	// acquire throttle for both src and tgt to avoid deadlocks
	tList := []*pqueue.Queue[reqmeta.Data]{}
	schemeSrcAPI, err := rc.schemeGet(refSrc.Scheme)
	if err != nil {
		return err
	}
	schemeTgtAPI, err := rc.schemeGet(refTgt.Scheme)
	if err != nil {
		return err
	}
	if tSrc, ok := schemeSrcAPI.(scheme.Throttler); ok {
		tList = append(tList, tSrc.Throttle(refSrc, false)...)
	}
	if tTgt, ok := schemeTgtAPI.(scheme.Throttler); ok {
		tList = append(tList, tTgt.Throttle(refTgt, true)...)
	}
	if len(tList) > 0 {
		ctxMulti, done, err := pqueue.AcquireMulti[reqmeta.Data](ctx, reqmeta.Data{Kind: reqmeta.Blob, Size: d.Size}, tList...)
		if err != nil {
			return err
		}
		if done != nil {
			defer done()
		}
		ctx = ctxMulti
	}

	// try mounting blob from the source repo is the registry is the same
	if ref.EqualRegistry(refSrc, refTgt) {
		err := rc.BlobMount(ctx, refSrc, refTgt, d)
		if err == nil {
			if opt.callback != nil {
				opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackSkipped, 0, d.Size)
			}
			rc.slog.Debug("Blob copy performed server side with registry mount",
				slog.String("src", refSrc.Reference),
				slog.String("tgt", refTgt.Reference),
				slog.String("digest", string(d.Digest)))
			return nil
		}
		rc.slog.Warn("Failed to mount blob",
			slog.String("src", refSrc.Reference),
			slog.String("tgt", refTgt.Reference),
			slog.String("err", err.Error()))
	}
	// fast options failed, download layer from source and push to target
	blobIO, err := rc.BlobGet(ctx, refSrc, d)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			rc.slog.Warn("Failed to retrieve blob",
				slog.String("src", refSrc.Reference),
				slog.String("digest", string(d.Digest)),
				slog.String("err", err.Error()))
		}
		return err
	}
	if opt.callback != nil {
		opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackStarted, 0, d.Size)
		ticker := time.NewTicker(blobCBFreq)
		done := make(chan bool)
		defer func() {
			close(done)
			ticker.Stop()
			if ctx.Err() == nil {
				opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackFinished, d.Size, d.Size)
			}
		}()
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					offset, err := blobIO.Seek(0, io.SeekCurrent)
					if err == nil && offset > 0 {
						opt.callback(types.CallbackBlob, d.Digest.String(), types.CallbackActive, offset, d.Size)
					}
				}
			}
		}()
	}
	if opt.readerHook != nil {
		blobIO, err = opt.readerHook(blobIO)
		if err != nil {
			rc.slog.Warn("Failed to apply reader hook to blob",
				slog.String("src", refSrc.Reference),
				slog.String("err", err.Error()))
			return err
		}
	}
	defer blobIO.Close()
	if _, err := rc.BlobPut(ctx, refTgt, blobIO.GetDescriptor(), blobIO); err != nil {
		if !errors.Is(err, context.Canceled) {
			rc.slog.Warn("Failed to push blob",
				slog.String("src", refSrc.Reference),
				slog.String("tgt", refTgt.Reference),
				slog.String("err", err.Error()))
		}
		return err
	}
	return nil
}

// BlobDelete removes a blob from the registry.
// This method should only be used to repair a damaged registry.
// Typically a server side garbage collection should be used to purge unused blobs.
func (rc *RegClient) BlobDelete(ctx context.Context, r ref.Ref, d descriptor.Descriptor) error {
	if !r.IsSetRepo() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return err
	}
	return schemeAPI.BlobDelete(ctx, r, d)
}

// BlobGet retrieves a blob, returning a reader.
// This reader must be closed to free up resources that limit concurrent pulls.
func (rc *RegClient) BlobGet(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error) {
	data, err := d.GetData()
	if err == nil {
		return blob.NewReader(blob.WithDesc(d), blob.WithRef(r), blob.WithReader(bytes.NewReader(data))), nil
	}
	if !r.IsSetRepo() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return nil, err
	}
	return schemeAPI.BlobGet(ctx, r, d)
}

// BlobGetOCIConfig retrieves an OCI config from a blob, automatically extracting the JSON.
func (rc *RegClient) BlobGetOCIConfig(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.OCIConfig, error) {
	if !r.IsSetRepo() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	b, err := rc.BlobGet(ctx, r, d)
	if err != nil {
		return nil, err
	}
	return b.ToOCIConfig()
}

// BlobHead is used to verify if a blob exists and is accessible.
func (rc *RegClient) BlobHead(ctx context.Context, r ref.Ref, d descriptor.Descriptor) (blob.Reader, error) {
	if !r.IsSetRepo() {
		return nil, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return nil, err
	}
	return schemeAPI.BlobHead(ctx, r, d)
}

// BlobMount attempts to perform a server side copy/mount of the blob between repositories.
func (rc *RegClient) BlobMount(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor) error {
	if !refSrc.IsSetRepo() {
		return fmt.Errorf("ref is not set: %s%.0w", refSrc.CommonName(), errs.ErrInvalidReference)
	}
	if !refTgt.IsSetRepo() {
		return fmt.Errorf("ref is not set: %s%.0w", refTgt.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(refSrc.Scheme)
	if err != nil {
		return err
	}
	return schemeAPI.BlobMount(ctx, refSrc, refTgt, d)
}

// BlobPut uploads a blob to a repository.
// Descriptor is optional, leave size and digest to zero value if unknown.
// Reader must also be an [io.Seeker] to support chunked upload fallback.
//
// This will attempt an anonymous blob mount first which some registries may support.
// It will then try doing a full put of the blob without chunking (most widely supported).
// If the full put fails, it will fall back to a chunked upload (useful for flaky networks).
func (rc *RegClient) BlobPut(ctx context.Context, r ref.Ref, d descriptor.Descriptor, rdr io.Reader) (descriptor.Descriptor, error) {
	if !r.IsSetRepo() {
		return descriptor.Descriptor{}, fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return descriptor.Descriptor{}, err
	}
	return schemeAPI.BlobPut(ctx, r, d, rdr)
}
