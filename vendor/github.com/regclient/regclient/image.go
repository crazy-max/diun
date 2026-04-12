package regclient

import (
	"archive/tar"
	"cmp"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	// crypto libraries included for go-digest
	_ "crypto/sha256"
	_ "crypto/sha512"

	digest "github.com/opencontainers/go-digest"

	"github.com/regclient/regclient/pkg/archive"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker/schema2"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/warning"
)

const (
	dockerManifestFilename = "manifest.json"
	ociLayoutVersion       = "1.0.0"
	ociIndexFilename       = "index.json"
	ociLayoutFilename      = "oci-layout"
	annotationRefName      = "org.opencontainers.image.ref.name"
	annotationImageName    = "io.containerd.image.name"
)

// used by import/export to match docker tar expected format
type dockerTarManifest struct {
	Config       string
	RepoTags     []string
	Layers       []string
	Parent       digest.Digest                           `json:",omitempty"`
	LayerSources map[digest.Digest]descriptor.Descriptor `json:",omitempty"`
}

type (
	tarFileHandler func(header *tar.Header, trd *tarReadData) error
	tarReadData    struct {
		tr          *tar.Reader
		name        string
		handleAdded bool
		handlers    map[string]tarFileHandler
		links       map[string][]string
		processed   map[string]bool
		finish      []func() error
		// data processed from various handlers
		manifests           map[digest.Digest]manifest.Manifest
		ociIndex            v1.Index
		ociManifest         manifest.Manifest
		dockerManifestFound bool
		dockerManifestList  []dockerTarManifest
		dockerManifest      schema2.Manifest
	}
)

type tarWriteData struct {
	tw    *tar.Writer
	dirs  map[string]bool
	files map[string]bool
	// uid, gid  int
	mode      int64
	timestamp time.Time
}

type imageOpt struct {
	callback        func(kind types.CallbackKind, instance string, state types.CallbackState, cur, total int64)
	checkBaseDigest string
	checkBaseRef    string
	checkSkipConfig bool
	child           bool
	exportCompress  bool
	exportRef       ref.Ref
	fastCheck       bool
	forceRecursive  bool
	importName      string
	includeExternal bool
	digestTags      bool
	platform        string
	platforms       []string
	referrerConfs   []scheme.ReferrerConfig
	referrerSrc     ref.Ref
	referrerTgt     ref.Ref
	tagList         []string
	mu              sync.Mutex
	seen            map[string]*imageSeen
	finalFn         []func(context.Context) error
	blobReaderHook  func(*blob.BReader) (*blob.BReader, error)
}

type imageSeen struct {
	done chan struct{}
	err  error
}

// ImageOpts define options for the Image* commands.
type ImageOpts func(*imageOpt)

// ImageWithBlobReaderHook calls the given function on every blob copy in [RegClient.ImageCopy].
// The hook receives a [blob.BReader] from getting the blob from the source.
// The returned [blob.BReader] will be used for pushing the blob to the target.
// If the hook returns an error on any blob, the image copy may fail.
func ImageWithBlobReaderHook(fn func(*blob.BReader) (*blob.BReader, error)) ImageOpts {
	return func(opts *imageOpt) {
		opts.blobReaderHook = fn
	}
}

// ImageWithCallback provides progress data to a callback function.
func ImageWithCallback(callback func(kind types.CallbackKind, instance string, state types.CallbackState, cur, total int64)) ImageOpts {
	return func(opts *imageOpt) {
		opts.callback = callback
	}
}

// ImageWithCheckBaseDigest provides a base digest to compare in ImageCheckBase.
func ImageWithCheckBaseDigest(d string) ImageOpts {
	return func(opts *imageOpt) {
		opts.checkBaseDigest = d
	}
}

// ImageWithCheckBaseRef provides a base reference to compare in ImageCheckBase.
func ImageWithCheckBaseRef(r string) ImageOpts {
	return func(opts *imageOpt) {
		opts.checkBaseRef = r
	}
}

// ImageWithCheckSkipConfig skips the configuration check in ImageCheckBase.
func ImageWithCheckSkipConfig() ImageOpts {
	return func(opts *imageOpt) {
		opts.checkSkipConfig = true
	}
}

// ImageWithChild attempts to copy every manifest and blob even if parent manifests already exist in ImageCopy.
func ImageWithChild() ImageOpts {
	return func(opts *imageOpt) {
		opts.child = true
	}
}

// ImageWithExportCompress adds gzip compression to tar export output in ImageExport.
func ImageWithExportCompress() ImageOpts {
	return func(opts *imageOpt) {
		opts.exportCompress = true
	}
}

// ImageWithExportRef overrides the image name embedded in the export file in ImageExport.
func ImageWithExportRef(r ref.Ref) ImageOpts {
	return func(opts *imageOpt) {
		opts.exportRef = r
	}
}

// ImageWithFastCheck skips check for referrers when manifest has already been copied in ImageCopy.
func ImageWithFastCheck() ImageOpts {
	return func(opts *imageOpt) {
		opts.fastCheck = true
	}
}

// ImageWithForceRecursive attempts to copy every manifest and blob even if parent manifests already exist in ImageCopy.
func ImageWithForceRecursive() ImageOpts {
	return func(opts *imageOpt) {
		opts.forceRecursive = true
	}
}

// ImageWithImportName selects the name of the image to import when multiple images are included in ImageImport.
func ImageWithImportName(name string) ImageOpts {
	return func(opts *imageOpt) {
		opts.importName = name
	}
}

// ImageWithIncludeExternal attempts to copy every manifest and blob even if parent manifests already exist in ImageCopy.
func ImageWithIncludeExternal() ImageOpts {
	return func(opts *imageOpt) {
		opts.includeExternal = true
	}
}

// ImageWithDigestTags looks for "sha-<digest>.*" tags in the repo to copy with any manifest in ImageCopy.
// These are used by some artifact systems like sigstore/cosign.
func ImageWithDigestTags() ImageOpts {
	return func(opts *imageOpt) {
		opts.digestTags = true
	}
}

// ImageWithPlatform requests specific platforms from a manifest list in ImageCheckBase.
func ImageWithPlatform(p string) ImageOpts {
	return func(opts *imageOpt) {
		opts.platform = p
	}
}

// ImageWithPlatforms only copies specific platforms from a manifest list in ImageCopy.
// This will result in a failure on many registries that validate manifests.
// Use the empty string to indicate images without a platform definition should be copied.
func ImageWithPlatforms(p []string) ImageOpts {
	return func(opts *imageOpt) {
		opts.platforms = p
	}
}

// ImageWithReferrers recursively recursively includes referrer images in ImageCopy.
func ImageWithReferrers(rOpts ...scheme.ReferrerOpts) ImageOpts {
	return func(opts *imageOpt) {
		if opts.referrerConfs == nil {
			opts.referrerConfs = []scheme.ReferrerConfig{}
		}
		rConf := scheme.ReferrerConfig{}
		for _, rOpt := range rOpts {
			rOpt(&rConf)
		}
		opts.referrerConfs = append(opts.referrerConfs, rConf)
	}
}

// ImageWithReferrerSrc specifies an alternate repository to pull referrers from.
func ImageWithReferrerSrc(src ref.Ref) ImageOpts {
	return func(opts *imageOpt) {
		opts.referrerSrc = src
	}
}

// ImageWithReferrerTgt specifies an alternate repository to pull referrers from.
func ImageWithReferrerTgt(tgt ref.Ref) ImageOpts {
	return func(opts *imageOpt) {
		opts.referrerTgt = tgt
	}
}

// ImageCheckBase returns nil if the base image is unchanged.
// A base image mismatch returns an error that wraps errs.ErrMismatch.
func (rc *RegClient) ImageCheckBase(ctx context.Context, r ref.Ref, opts ...ImageOpts) error {
	var opt imageOpt
	for _, optFn := range opts {
		optFn(&opt)
	}
	var m manifest.Manifest
	var err error

	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// if the base name is not provided, check image for base annotations
	if opt.checkBaseRef == "" {
		m, err = rc.ManifestGet(ctx, r)
		if err != nil {
			return err
		}
		ma, ok := m.(manifest.Annotator)
		if !ok {
			return fmt.Errorf("image does not support annotations, base image must be provided%.0w", errs.ErrMissingAnnotation)
		}
		annot, err := ma.GetAnnotations()
		if err != nil {
			return err
		}
		if baseName, ok := annot[types.AnnotationBaseImageName]; ok {
			opt.checkBaseRef = baseName
		} else {
			return fmt.Errorf("image does not have a base annotation, base image must be provided%.0w", errs.ErrMissingAnnotation)
		}
		if baseDig, ok := annot[types.AnnotationBaseImageDigest]; ok {
			opt.checkBaseDigest = baseDig
		}
	}
	baseR, err := ref.New(opt.checkBaseRef)
	if err != nil {
		return err
	}
	defer rc.Close(ctx, baseR)

	// if the digest is available, check if that matches the base name
	if opt.checkBaseDigest != "" {
		baseMH, err := rc.ManifestHead(ctx, baseR, WithManifestRequireDigest())
		if err != nil {
			return err
		}
		expectDig, err := digest.Parse(opt.checkBaseDigest)
		if err != nil {
			return err
		}
		if baseMH.GetDescriptor().Digest == expectDig {
			rc.slog.Debug("base image digest matches",
				slog.String("name", baseR.CommonName()),
				slog.String("digest", baseMH.GetDescriptor().Digest.String()))
			return nil
		} else {
			rc.slog.Debug("base image digest changed",
				slog.String("name", baseR.CommonName()),
				slog.String("digest", baseMH.GetDescriptor().Digest.String()),
				slog.String("expected", expectDig.String()))
			return fmt.Errorf("base digest changed, %s, expected %s, received %s%.0w",
				baseR.CommonName(), expectDig.String(), baseMH.GetDescriptor().Digest.String(), errs.ErrMismatch)
		}
	}

	// if the digest is not available, compare layers of each manifest
	if m == nil {
		m, err = rc.ManifestGet(ctx, r)
		if err != nil {
			return err
		}
	}
	if m.IsList() && opt.platform != "" {
		p, err := platform.Parse(opt.platform)
		if err != nil {
			return err
		}
		d, err := manifest.GetPlatformDesc(m, &p)
		if err != nil {
			return err
		}
		rp := r.AddDigest(d.Digest.String())
		m, err = rc.ManifestGet(ctx, rp)
		if err != nil {
			return err
		}
	}
	if m.IsList() {
		// loop through each platform
		ml, ok := m.(manifest.Indexer)
		if !ok {
			return fmt.Errorf("manifest list is not an Indexer")
		}
		dl, err := ml.GetManifestList()
		if err != nil {
			return err
		}
		for _, d := range dl {
			rp := r.AddDigest(d.Digest.String())
			optP := append(opts, ImageWithPlatform(d.Platform.String()))
			err = rc.ImageCheckBase(ctx, rp, optP...)
			if err != nil {
				return fmt.Errorf("platform %s mismatch: %w", d.Platform.String(), err)
			}
		}
		return nil
	}
	img, ok := m.(manifest.Imager)
	if !ok {
		return fmt.Errorf("manifest must be an image")
	}
	layers, err := img.GetLayers()
	if err != nil {
		return err
	}
	baseM, err := rc.ManifestGet(ctx, baseR)
	if err != nil {
		return err
	}
	if baseM.IsList() && opt.platform != "" {
		p, err := platform.Parse(opt.platform)
		if err != nil {
			return err
		}
		d, err := manifest.GetPlatformDesc(baseM, &p)
		if err != nil {
			return err
		}
		baseM, err = rc.ManifestGet(ctx, baseR, WithManifestDesc(*d))
		if err != nil {
			return err
		}
	}
	baseImg, ok := baseM.(manifest.Imager)
	if !ok {
		return fmt.Errorf("base image manifest must be an image")
	}
	baseLayers, err := baseImg.GetLayers()
	if err != nil {
		return err
	}
	if len(baseLayers) <= 0 {
		return fmt.Errorf("base image has no layers")
	}
	for i := range baseLayers {
		if i >= len(layers) {
			return fmt.Errorf("image has fewer layers than base image")
		}
		if !layers[i].Same(baseLayers[i]) {
			rc.slog.Debug("image layer changed",
				slog.Int("layer", i),
				slog.String("expected", layers[i].Digest.String()),
				slog.String("digest", baseLayers[i].Digest.String()))
			return fmt.Errorf("base layer changed, %s[%d], expected %s, received %s%.0w",
				baseR.CommonName(), i, layers[i].Digest.String(), baseLayers[i].Digest.String(), errs.ErrMismatch)
		}
	}

	if opt.checkSkipConfig {
		return nil
	}

	// if the layers match, compare the config history
	confDesc, err := img.GetConfig()
	if err != nil {
		return err
	}
	conf, err := rc.BlobGetOCIConfig(ctx, r, confDesc)
	if err != nil {
		return err
	}
	confOCI := conf.GetConfig()
	baseConfDesc, err := baseImg.GetConfig()
	if err != nil {
		return err
	}
	baseConf, err := rc.BlobGetOCIConfig(ctx, baseR, baseConfDesc)
	if err != nil {
		return err
	}
	baseConfOCI := baseConf.GetConfig()
	for i := range baseConfOCI.History {
		if i >= len(confOCI.History) {
			return fmt.Errorf("image has fewer history entries than base image")
		}
		if baseConfOCI.History[i].Author != confOCI.History[i].Author ||
			baseConfOCI.History[i].Comment != confOCI.History[i].Comment ||
			!baseConfOCI.History[i].Created.Equal(*confOCI.History[i].Created) ||
			baseConfOCI.History[i].CreatedBy != confOCI.History[i].CreatedBy ||
			baseConfOCI.History[i].EmptyLayer != confOCI.History[i].EmptyLayer {
			rc.slog.Debug("image history changed",
				slog.Int("index", i),
				slog.Any("expected", confOCI.History[i]),
				slog.Any("history", baseConfOCI.History[i]))
			return fmt.Errorf("base history changed, %s[%d], expected %v, received %v%.0w",
				baseR.CommonName(), i, confOCI.History[i], baseConfOCI.History[i], errs.ErrMismatch)
		}
	}

	rc.slog.Debug("base image layers and history matches",
		slog.String("base", baseR.CommonName()))
	return nil
}

// ImageConfig returns the OCI config of a given image.
// Use [ImageWithPlatform] to select a platform from an Index or Manifest List.
func (rc *RegClient) ImageConfig(ctx context.Context, r ref.Ref, opts ...ImageOpts) (*blob.BOCIConfig, error) {
	opt := imageOpt{
		platform: "local",
	}
	for _, optFn := range opts {
		optFn(&opt)
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	p, err := platform.Parse(opt.platform)
	if err != nil {
		return nil, fmt.Errorf("failed to parse platform %s: %w", opt.platform, err)
	}
	m, err := rc.ManifestGet(ctx, r, WithManifestPlatform(p))
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	for m.IsList() {
		mi, ok := m.(manifest.Indexer)
		if !ok {
			return nil, fmt.Errorf("unsupported manifest type: %s", m.GetDescriptor().MediaType)
		}
		ml, err := mi.GetManifestList()
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest list: %w", err)
		}
		d, err := descriptor.DescriptorListSearch(ml, descriptor.MatchOpt{Platform: &p})
		if err != nil {
			return nil, fmt.Errorf("failed to find platform in manifest list: %w", err)
		}
		m, err = rc.ManifestGet(ctx, r, WithManifestDesc(d))
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest: %w", err)
		}
	}
	mi, ok := m.(manifest.Imager)
	if !ok {
		return nil, fmt.Errorf("unsupported manifest type: %s", m.GetDescriptor().MediaType)
	}
	d, err := mi.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get image config: %w", err)
	}
	if d.MediaType != mediatype.OCI1ImageConfig && d.MediaType != mediatype.Docker2ImageConfig {
		return nil, fmt.Errorf("unsupported config media type %s: %w", d.MediaType, errs.ErrUnsupportedMediaType)
	}
	return rc.BlobGetOCIConfig(ctx, r, d)
}

// ImageCopy copies an image.
// This will retag an image in the same repository, only pushing and pulling the top level manifest.
// On the same registry, it will attempt to use cross-repository blob mounts to avoid pulling blobs.
// Blobs are only pulled when they don't exist on the target and a blob mount fails.
// Referrers are optionally copied recursively.
func (rc *RegClient) ImageCopy(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, opts ...ImageOpts) error {
	opt := imageOpt{
		seen:    map[string]*imageSeen{},
		finalFn: []func(context.Context) error{},
	}
	for _, optFn := range opts {
		optFn(&opt)
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// block GC from running (in OCIDir) during the copy
	schemeTgtAPI, err := rc.schemeGet(refTgt.Scheme)
	if err != nil {
		return err
	}
	if tgtGCLocker, isGCLocker := schemeTgtAPI.(scheme.GCLocker); isGCLocker {
		tgtGCLocker.GCLock(refTgt)
		defer tgtGCLocker.GCUnlock(refTgt)
	}
	// run the copy of manifests and blobs recursively
	err = rc.imageCopyOpt(ctx, refSrc, refTgt, descriptor.Descriptor{}, opt.child, []digest.Digest{}, &opt)
	if err != nil {
		return err
	}
	// run any final functions, digest-tags and referrers that detected loops are retried here
	for _, fn := range opt.finalFn {
		err := fn(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// imageCopyOpt is a thread safe copy of a manifest and nested content.
func (rc *RegClient) imageCopyOpt(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor, child bool, parents []digest.Digest, opt *imageOpt) (err error) {
	var mSrc, mTgt manifest.Manifest
	var sDig digest.Digest
	refTgtRepo := refTgt.SetTag("").CommonName()
	seenCB := func(error) {}
	defer func() {
		if seenCB != nil {
			seenCB(err)
		}
	}()
	// if digest is provided and we are already copying it, wait
	if d.Digest != "" {
		sDig = d.Digest
	} else if refSrc.Digest != "" {
		sDig = digest.Digest(refSrc.Digest)
	}
	if sDig != "" {
		if seenCB, err = imageSeenOrWait(ctx, opt, refTgtRepo, refTgt.Tag, sDig, parents); seenCB == nil {
			return err
		}
	}
	// check target with head request
	mTgt, err = rc.ManifestHead(ctx, refTgt, WithManifestRequireDigest())
	var urlError *url.Error
	if err != nil && errors.As(err, &urlError) {
		return fmt.Errorf("failed to access target registry: %w", err)
	}
	// for non-recursive copies, compare to source digest
	if err == nil && (opt.fastCheck || (!opt.forceRecursive && opt.referrerConfs == nil && !opt.digestTags)) {
		if sDig == "" {
			mSrc, err = rc.ManifestHead(ctx, refSrc, WithManifestRequireDigest())
			if err != nil {
				return fmt.Errorf("copy failed, error getting source: %w", err)
			}
			sDig = mSrc.GetDescriptor().Digest
			if seenCB, err = imageSeenOrWait(ctx, opt, refTgtRepo, refTgt.Tag, sDig, parents); seenCB == nil {
				return err
			}
		}
		if sDig == mTgt.GetDescriptor().Digest {
			if opt.callback != nil {
				opt.callback(types.CallbackManifest, d.Digest.String(), types.CallbackSkipped, mTgt.GetDescriptor().Size, mTgt.GetDescriptor().Size)
			}
			return nil
		}
	}
	// when copying/updating digest tags or referrers, only the source digest is needed for an image
	if mTgt != nil && mSrc == nil && !opt.forceRecursive && sDig == "" {
		mSrc, err = rc.ManifestHead(ctx, refSrc, WithManifestRequireDigest())
		if err != nil {
			return fmt.Errorf("copy failed, error getting source: %w", err)
		}
		sDig = mSrc.GetDescriptor().Digest
		if seenCB, err = imageSeenOrWait(ctx, opt, refTgtRepo, refTgt.Tag, sDig, parents); seenCB == nil {
			return err
		}
	}
	// get the source manifest when a copy is needed or recursion into the content is needed
	if sDig == "" || mTgt == nil || sDig != mTgt.GetDescriptor().Digest || opt.forceRecursive || mTgt.IsList() {
		mSrc, err = rc.ManifestGet(ctx, refSrc, WithManifestDesc(d))
		if err != nil {
			return fmt.Errorf("copy failed, error getting source: %w", err)
		}
		if sDig == "" {
			sDig = mSrc.GetDescriptor().Digest
			if seenCB, err = imageSeenOrWait(ctx, opt, refTgtRepo, refTgt.Tag, sDig, parents); seenCB == nil {
				return err
			}
		}
	}
	// setup vars for a copy
	mOpts := []ManifestOpts{}
	if child {
		mOpts = append(mOpts, WithManifestChild())
	}
	bOpt := []BlobOpts{}
	if opt.callback != nil {
		bOpt = append(bOpt, BlobWithCallback(opt.callback))
	}
	if opt.blobReaderHook != nil {
		bOpt = append(bOpt, BlobWithReaderHook(opt.blobReaderHook))
	}
	waitCh := make(chan error)
	waitCount := 0
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	parentsNew := make([]digest.Digest, len(parents)+1)
	copy(parentsNew, parents)
	parentsNew[len(parentsNew)-1] = sDig
	if opt.callback != nil {
		opt.callback(types.CallbackManifest, d.Digest.String(), types.CallbackStarted, 0, d.Size)
	}
	// process entries in an index
	if mSrcIndex, ok := mSrc.(manifest.Indexer); ok && mSrc.IsSet() && !ref.EqualRepository(refSrc, refTgt) {
		// manifest lists need to recursively copy nested images by digest
		dList, err := mSrcIndex.GetManifestList()
		if err != nil {
			return err
		}
		for _, dEntry := range dList {
			// skip copy of platforms not specifically included
			if len(opt.platforms) > 0 {
				match, err := imagePlatformInList(dEntry.Platform, opt.platforms)
				if err != nil {
					return err
				}
				if !match {
					rc.slog.Debug("Platform excluded from copy",
						slog.Any("platform", dEntry.Platform))
					continue
				}
			}
			waitCount++
			go func() {
				var err error
				rc.slog.Debug("Copy platform",
					slog.Any("platform", dEntry.Platform),
					slog.String("digest", dEntry.Digest.String()))
				entrySrc := refSrc.SetDigest(dEntry.Digest.String())
				entryTgt := refTgt.SetDigest(dEntry.Digest.String())
				switch dEntry.MediaType {
				case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned,
					mediatype.Docker2Manifest, mediatype.Docker2ManifestList,
					mediatype.OCI1Manifest, mediatype.OCI1ManifestList:
					// known manifest media type
					err = rc.imageCopyOpt(ctx, entrySrc, entryTgt, dEntry, true, parentsNew, opt)
				case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
					mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
					mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
					mediatype.BuildkitCacheConfig:
					// known blob media type
					err = rc.imageCopyBlob(ctx, entrySrc, entryTgt, dEntry, opt, bOpt...)
				default:
					// unknown media type, first try an image copy
					err = rc.imageCopyOpt(ctx, entrySrc, entryTgt, dEntry, true, parentsNew, opt)
					if err != nil {
						// fall back to trying to copy a blob
						err = rc.imageCopyBlob(ctx, entrySrc, entryTgt, dEntry, opt, bOpt...)
					}
				}
				waitCh <- err
			}()
		}
	}

	// If source is image, copy blobs
	if mSrcImg, ok := mSrc.(manifest.Imager); ok && mSrc.IsSet() && !ref.EqualRepository(refSrc, refTgt) {
		// copy the config
		cd, err := mSrcImg.GetConfig()
		if err != nil {
			// docker schema v1 does not have a config object, ignore if it's missing
			if !errors.Is(err, errs.ErrUnsupportedMediaType) {
				rc.slog.Warn("Failed to get config digest from manifest",
					slog.String("ref", refSrc.Reference),
					slog.String("err", err.Error()))
				return fmt.Errorf("failed to get config digest for %s: %w", refSrc.CommonName(), err)
			}
		} else {
			waitCount++
			go func() {
				rc.slog.Info("Copy config",
					slog.String("source", refSrc.Reference),
					slog.String("target", refTgt.Reference),
					slog.String("digest", cd.Digest.String()))
				err := rc.imageCopyBlob(ctx, refSrc, refTgt, cd, opt, bOpt...)
				if err != nil && !errors.Is(err, context.Canceled) {
					rc.slog.Warn("Failed to copy config",
						slog.String("source", refSrc.Reference),
						slog.String("target", refTgt.Reference),
						slog.String("digest", cd.Digest.String()),
						slog.String("err", err.Error()))
				}
				waitCh <- err
			}()
		}

		// copy filesystem layers
		l, err := mSrcImg.GetLayers()
		if err != nil {
			return err
		}
		for _, layerSrc := range l {
			if len(layerSrc.URLs) > 0 && !opt.includeExternal {
				// skip blobs where the URLs are defined, these aren't hosted and won't be pulled from the source
				rc.slog.Debug("Skipping external layer",
					slog.String("source", refSrc.Reference),
					slog.String("target", refTgt.Reference),
					slog.String("layer", layerSrc.Digest.String()),
					slog.Any("external-urls", layerSrc.URLs))
				continue
			}
			waitCount++
			go func() {
				rc.slog.Info("Copy layer",
					slog.String("source", refSrc.Reference),
					slog.String("target", refTgt.Reference),
					slog.String("layer", layerSrc.Digest.String()))
				err := rc.imageCopyBlob(ctx, refSrc, refTgt, layerSrc, opt, bOpt...)
				if err != nil && !errors.Is(err, context.Canceled) {
					rc.slog.Warn("Failed to copy layer",
						slog.String("source", refSrc.Reference),
						slog.String("target", refTgt.Reference),
						slog.String("layer", layerSrc.Digest.String()),
						slog.String("err", err.Error()))
				}
				waitCh <- err
			}()
		}
	}

	// check for any errors and abort early if found
	err = nil
	done := false
	for !done && waitCount > 0 {
		if err == nil {
			select {
			case err = <-waitCh:
				if err != nil {
					cancel()
				}
			default:
				done = true // happy path
			}
		} else {
			if errors.Is(err, context.Canceled) {
				// try to find a better error message than context canceled
				err = <-waitCh
			} else {
				<-waitCh
			}
		}
		if !done {
			waitCount--
		}
	}
	if err != nil {
		rc.slog.Debug("child manifest copy failed",
			slog.String("err", err.Error()),
			slog.String("sDig", sDig.String()))
		return err
	}

	// copy referrers
	referrerTags := []string{}
	if opt.referrerConfs != nil {
		referrerOpts := []scheme.ReferrerOpts{}
		rSubject := refSrc
		referrerSrc := refSrc
		referrerTgt := refTgt
		if opt.referrerSrc.IsSet() {
			referrerOpts = append(referrerOpts, scheme.WithReferrerSource(opt.referrerSrc))
			referrerSrc = opt.referrerSrc
		}
		if opt.referrerTgt.IsSet() {
			referrerTgt = opt.referrerTgt
		}
		if sDig != "" {
			rSubject = rSubject.SetDigest(sDig.String())
		}
		rl, err := rc.ReferrerList(ctx, rSubject, referrerOpts...)
		if err != nil {
			return err
		}
		if !rl.Source.IsSet() || ref.EqualRepository(refSrc, rl.Source) {
			referrerTags = append(referrerTags, rl.Tags...)
		}
		descList := []descriptor.Descriptor{}
		if len(opt.referrerConfs) == 0 {
			descList = rl.Descriptors
		} else {
			for _, rConf := range opt.referrerConfs {
				rlFilter := scheme.ReferrerFilter(rConf, rl)
				descList = append(descList, rlFilter.Descriptors...)
			}
		}
		for _, rDesc := range descList {
			opt.mu.Lock()
			seen := opt.seen[":"+rDesc.Digest.String()]
			opt.mu.Unlock()
			if seen != nil {
				continue // skip referrers that have been seen
			}
			referrerSrc := referrerSrc.SetDigest(rDesc.Digest.String())
			referrerTgt := referrerTgt.SetDigest(rDesc.Digest.String())
			waitCount++
			go func() {
				err := rc.imageCopyOpt(ctx, referrerSrc, referrerTgt, rDesc, true, parentsNew, opt)
				if errors.Is(err, errs.ErrLoopDetected) {
					// if a loop is detected, push the referrers copy to the end
					opt.mu.Lock()
					opt.finalFn = append(opt.finalFn, func(ctx context.Context) error {
						return rc.imageCopyOpt(ctx, referrerSrc, referrerTgt, rDesc, true, []digest.Digest{}, opt)
					})
					opt.mu.Unlock()
					waitCh <- nil
				} else {
					if err != nil && !errors.Is(err, context.Canceled) {
						rc.slog.Warn("Failed to copy referrer",
							slog.String("digest", rDesc.Digest.String()),
							slog.String("src", referrerSrc.CommonName()),
							slog.String("tgt", referrerTgt.CommonName()))
					}
					waitCh <- err
				}
			}()
		}
	}

	// lookup digest tags to include artifacts with image
	if opt.digestTags {
		// load tag listing for digest tag copy
		opt.mu.Lock()
		if opt.tagList == nil {
			tl, err := rc.TagList(ctx, refSrc)
			if err != nil {
				opt.mu.Unlock()
				rc.slog.Warn("Failed to list tags for digest-tag copy",
					slog.String("source", refSrc.Reference),
					slog.String("err", err.Error()))
				return err
			}
			tags, err := tl.GetTags()
			if err != nil {
				opt.mu.Unlock()
				rc.slog.Warn("Failed to list tags for digest-tag copy",
					slog.String("source", refSrc.Reference),
					slog.String("err", err.Error()))
				return err
			}
			if tags == nil {
				tags = []string{}
			}
			opt.tagList = tags
		}
		opt.mu.Unlock()
		prefix := fmt.Sprintf("%s-%s", sDig.Algorithm(), sDig.Encoded())
		for _, tag := range opt.tagList {
			if strings.HasPrefix(tag, prefix) {
				// skip referrers that were copied above
				if slices.Contains(referrerTags, tag) {
					continue
				}
				refTagSrc := refSrc.SetTag(tag)
				refTagTgt := refTgt.SetTag(tag)
				waitCount++
				go func() {
					err := rc.imageCopyOpt(ctx, refTagSrc, refTagTgt, descriptor.Descriptor{}, false, parentsNew, opt)
					if errors.Is(err, errs.ErrLoopDetected) {
						// if a loop is detected, push the digest tag copy back to the end
						opt.mu.Lock()
						opt.finalFn = append(opt.finalFn, func(ctx context.Context) error {
							return rc.imageCopyOpt(ctx, refTagSrc, refTagTgt, descriptor.Descriptor{}, false, []digest.Digest{}, opt)
						})
						opt.mu.Unlock()
						waitCh <- nil
					} else {
						if err != nil && !errors.Is(err, context.Canceled) {
							rc.slog.Warn("Failed to copy digest-tag",
								slog.String("tag", tag),
								slog.String("src", refTagSrc.CommonName()),
								slog.String("tgt", refTagTgt.CommonName()))
						}
						waitCh <- err
					}
				}()
			}
		}
	}

	// wait for background tasks to finish
	err = nil
	for waitCount > 0 {
		if err == nil {
			err = <-waitCh
			if err != nil {
				cancel()
			}
		} else {
			if errors.Is(err, context.Canceled) {
				// try to find a better error message than context canceled
				err = <-waitCh
			} else {
				<-waitCh
			}
		}
		waitCount--
	}
	if err != nil {
		return err
	}

	// push manifest
	if mTgt == nil || sDig != mTgt.GetDescriptor().Digest || opt.forceRecursive {
		err = rc.ManifestPut(ctx, refTgt, mSrc, mOpts...)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				rc.slog.Warn("Failed to push manifest",
					slog.String("target", refTgt.Reference),
					slog.String("err", err.Error()))
			}
			return err
		}
		if opt.callback != nil {
			opt.callback(types.CallbackManifest, d.Digest.String(), types.CallbackFinished, d.Size, d.Size)
		}
	} else {
		if opt.callback != nil {
			opt.callback(types.CallbackManifest, d.Digest.String(), types.CallbackSkipped, d.Size, d.Size)
		}
	}
	if seenCB != nil {
		seenCB(nil)
		seenCB = nil
	}

	return nil
}

func (rc *RegClient) imageCopyBlob(ctx context.Context, refSrc ref.Ref, refTgt ref.Ref, d descriptor.Descriptor, opt *imageOpt, bOpt ...BlobOpts) error {
	seenCB, err := imageSeenOrWait(ctx, opt, refTgt.SetTag("").CommonName(), "", d.Digest, []digest.Digest{})
	if seenCB == nil {
		return err
	}
	err = rc.BlobCopy(ctx, refSrc, refTgt, d, bOpt...)
	seenCB(err)
	return err
}

// imageSeenOrWait returns either a callback to report the error when the digest hasn't been seen before
// or it will wait for the previous copy to run and return the error from that copy
func imageSeenOrWait(ctx context.Context, opt *imageOpt, repo, tag string, dig digest.Digest, parents []digest.Digest) (func(error), error) {
	var seenNew *imageSeen
	key := repo + "/" + tag + ":" + dig.String()
	opt.mu.Lock()
	seen := opt.seen[key]
	if seen == nil {
		seenNew = &imageSeen{
			done: make(chan struct{}),
		}
		opt.seen[key] = seenNew
	}
	opt.mu.Unlock()
	if seen != nil {
		// quick check for the previous copy already done
		select {
		case <-seen.done:
			return nil, seen.err
		default:
		}
		// look for loops in parents
		for _, p := range parents {
			if key == repo+"/"+tag+":"+p.String() {
				return nil, errs.ErrLoopDetected
			}
		}
		// wait for copy to finish or context to cancel
		done := ctx.Done()
		select {
		case <-seen.done:
			return nil, seen.err
		case <-done:
			return nil, ctx.Err()
		}
	} else {
		return func(err error) {
			seenNew.err = err
			close(seenNew.done)
			// on failures, delete the history to allow a retry
			if err != nil {
				opt.mu.Lock()
				delete(opt.seen, key)
				opt.mu.Unlock()
			}
		}, nil
	}
}

// ImageExport exports an image to an output stream.
// The format is compatible with "docker load" if a single image is selected and not a manifest list.
// The ref must include a tag for exporting to docker (defaults to latest), and may also include a digest.
// The export is also formatted according to [OCI Layout] which supports multi-platform images.
// A tar file will be sent to outStream.
//
// Resulting filesystem:
//   - oci-layout: created at top level, can be done at the start
//   - index.json: created at top level, single descriptor with org.opencontainers.image.ref.name annotation pointing to the tag
//   - manifest.json: created at top level, based on every layer added, only works for a single arch image
//   - blobs/$algo/$hash: each content addressable object (manifest, config, or layer), created recursively
//
// [OCI Layout]: https://github.com/opencontainers/image-spec/blob/master/image-layout.md
func (rc *RegClient) ImageExport(ctx context.Context, r ref.Ref, outStream io.Writer, opts ...ImageOpts) error {
	if !r.IsSet() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	var ociIndex v1.Index

	var opt imageOpt
	for _, optFn := range opts {
		optFn(&opt)
	}
	if opt.exportRef.IsZero() {
		opt.exportRef = r
	}

	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// create tar writer object
	out := outStream
	if opt.exportCompress {
		gzOut := gzip.NewWriter(out)
		defer gzOut.Close()
		out = gzOut
	}
	tw := tar.NewWriter(out)
	defer tw.Close()
	twd := &tarWriteData{
		tw:    tw,
		dirs:  map[string]bool{},
		files: map[string]bool{},
		mode:  0o644,
	}

	// retrieve image manifest
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
		rc.slog.Warn("Failed to get manifest",
			slog.String("ref", r.CommonName()),
			slog.String("err", err.Error()))
		return err
	}

	// build/write oci-layout
	ociLayout := v1.ImageLayout{Version: ociLayoutVersion}
	err = twd.tarWriteFileJSON(ociLayoutFilename, ociLayout)
	if err != nil {
		return err
	}

	// create a manifest descriptor
	mDesc := m.GetDescriptor()
	if mDesc.Annotations == nil {
		mDesc.Annotations = map[string]string{}
	}
	mDesc.Annotations[annotationImageName] = opt.exportRef.CommonName()
	mDesc.Annotations[annotationRefName] = opt.exportRef.Tag

	// generate/write an OCI index
	ociIndex.Versioned = v1.IndexSchemaVersion
	ociIndex.Manifests = []descriptor.Descriptor{mDesc} // initialize with the descriptor to the manifest list
	err = twd.tarWriteFileJSON(ociIndexFilename, ociIndex)
	if err != nil {
		return err
	}

	// append to docker manifest with tag, config filename, each layer filename, and layer descriptors
	if mi, ok := m.(manifest.Imager); ok {
		conf, err := mi.GetConfig()
		if err != nil {
			return err
		}
		if err = conf.Digest.Validate(); err != nil {
			return err
		}
		refTag := opt.exportRef.ToReg()
		refTag = refTag.SetTag(cmp.Or(refTag.Tag, "latest"))
		dockerManifest := dockerTarManifest{
			RepoTags:     []string{refTag.CommonName()},
			Config:       tarOCILayoutDescPath(conf),
			Layers:       []string{},
			LayerSources: map[digest.Digest]descriptor.Descriptor{},
		}
		dl, err := mi.GetLayers()
		if err != nil {
			return err
		}
		for _, d := range dl {
			if err = d.Digest.Validate(); err != nil {
				return err
			}
			dockerManifest.Layers = append(dockerManifest.Layers, tarOCILayoutDescPath(d))
			dockerManifest.LayerSources[d.Digest] = d
		}

		// marshal manifest and write manifest.json
		err = twd.tarWriteFileJSON(dockerManifestFilename, []dockerTarManifest{dockerManifest})
		if err != nil {
			return err
		}
	}

	// recursively include manifests and nested blobs
	err = rc.imageExportDescriptor(ctx, r, mDesc, twd)
	if err != nil {
		return err
	}

	return nil
}

// imageExportDescriptor pulls a manifest or blob, outputs to a tar file, and recursively processes any nested manifests or blobs
func (rc *RegClient) imageExportDescriptor(ctx context.Context, r ref.Ref, desc descriptor.Descriptor, twd *tarWriteData) error {
	if err := desc.Digest.Validate(); err != nil {
		return err
	}
	tarFilename := tarOCILayoutDescPath(desc)
	if twd.files[tarFilename] {
		// blob has already been imported into tar, skip
		return nil
	}
	switch desc.MediaType {
	case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned, mediatype.Docker2Manifest, mediatype.OCI1Manifest:
		// Handle single platform manifests
		// retrieve manifest
		m, err := rc.ManifestGet(ctx, r, WithManifestDesc(desc))
		if err != nil {
			return err
		}
		mi, ok := m.(manifest.Imager)
		if !ok {
			return fmt.Errorf("manifest doesn't support image methods%.0w", errs.ErrUnsupportedMediaType)
		}
		// write manifest body by digest
		mBody, err := m.RawBody()
		if err != nil {
			return err
		}
		err = twd.tarWriteHeader(tarFilename, int64(len(mBody)))
		if err != nil {
			return err
		}
		_, err = twd.tw.Write(mBody)
		if err != nil {
			return err
		}

		// add config
		confD, err := mi.GetConfig()
		// ignore unsupported media type errors
		if err != nil && !errors.Is(err, errs.ErrUnsupportedMediaType) {
			return err
		}
		if err == nil {
			err = rc.imageExportDescriptor(ctx, r, confD, twd)
			if err != nil {
				return err
			}
		}

		// loop over layers
		layerDL, err := mi.GetLayers()
		// ignore unsupported media type errors
		if err != nil && !errors.Is(err, errs.ErrUnsupportedMediaType) {
			return err
		}
		if err == nil {
			for _, layerD := range layerDL {
				err = rc.imageExportDescriptor(ctx, r, layerD, twd)
				if err != nil {
					return err
				}
			}
		}

	case mediatype.Docker2ManifestList, mediatype.OCI1ManifestList:
		// handle OCI index and Docker manifest list
		// retrieve manifest
		m, err := rc.ManifestGet(ctx, r, WithManifestDesc(desc))
		if err != nil {
			return err
		}
		mi, ok := m.(manifest.Indexer)
		if !ok {
			return fmt.Errorf("manifest doesn't support index methods%.0w", errs.ErrUnsupportedMediaType)
		}
		// write manifest body by digest
		mBody, err := m.RawBody()
		if err != nil {
			return err
		}
		err = twd.tarWriteHeader(tarFilename, int64(len(mBody)))
		if err != nil {
			return err
		}
		_, err = twd.tw.Write(mBody)
		if err != nil {
			return err
		}
		// recurse over entries in the list/index
		mdl, err := mi.GetManifestList()
		if err != nil {
			return err
		}
		for _, md := range mdl {
			err = rc.imageExportDescriptor(ctx, r, md, twd)
			if err != nil {
				return err
			}
		}

	default:
		// get blob
		blobR, err := rc.BlobGet(ctx, r, desc)
		if err != nil {
			return err
		}
		defer blobR.Close()
		// write blob by digest
		err = twd.tarWriteHeader(tarFilename, int64(desc.Size))
		if err != nil {
			return err
		}
		size, err := io.Copy(twd.tw, blobR)
		if err != nil {
			return fmt.Errorf("failed to export blob %s: %w", desc.Digest.String(), err)
		}
		if size != desc.Size {
			return fmt.Errorf("blob size mismatch, descriptor %d, received %d", desc.Size, size)
		}
	}

	return nil
}

// ImageImport pushes an image from a tar file (ImageExport) to a registry.
func (rc *RegClient) ImageImport(ctx context.Context, r ref.Ref, rs io.ReadSeeker, opts ...ImageOpts) error {
	if !r.IsSetRepo() {
		return fmt.Errorf("ref is not set: %s%.0w", r.CommonName(), errs.ErrInvalidReference)
	}
	var opt imageOpt
	for _, optFn := range opts {
		optFn(&opt)
	}

	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	trd := &tarReadData{
		name:      opt.importName,
		handlers:  map[string]tarFileHandler{},
		links:     map[string][]string{},
		processed: map[string]bool{},
		finish:    []func() error{},
		manifests: map[digest.Digest]manifest.Manifest{},
	}

	// add handler for oci-layout, index.json, and manifest.json
	rc.imageImportOCIAddHandler(ctx, r, trd)
	rc.imageImportDockerAddHandler(trd)

	// process tar file looking for oci-layout and index.json, load manifests/blobs on success
	err := trd.tarReadAll(rs)

	if err != nil && errors.Is(err, errs.ErrNotFound) && trd.dockerManifestFound {
		// import failed but manifest.json found, fall back to manifest.json processing
		// add handlers for the docker manifest layers
		rc.imageImportDockerAddLayerHandlers(ctx, r, trd)
		// reprocess the tar looking for manifest.json files
		err = trd.tarReadAll(rs)
		if err != nil {
			return fmt.Errorf("failed to import layers from docker tar: %w", err)
		}
		// push docker manifest
		m, err := manifest.New(manifest.WithOrig(trd.dockerManifest))
		if err != nil {
			return err
		}
		err = rc.ManifestPut(ctx, r, m)
		if err != nil {
			return err
		}
	} else if err != nil {
		// unhandled error from tar read
		return err
	} else {
		// successful load of OCI blobs, now push manifest and tag
		err = rc.imageImportOCIPushManifests(ctx, r, trd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *RegClient) imageImportBlob(ctx context.Context, r ref.Ref, desc descriptor.Descriptor, trd *tarReadData) error {
	// skip if blob already exists
	_, err := rc.BlobHead(ctx, r, desc)
	if err == nil {
		return nil
	}
	// upload blob
	_, err = rc.BlobPut(ctx, r, desc, trd.tr)
	if err != nil {
		return err
	}
	return nil
}

// imageImportDockerAddHandler processes tar files generated by docker.
func (rc *RegClient) imageImportDockerAddHandler(trd *tarReadData) {
	trd.handlers[dockerManifestFilename] = func(header *tar.Header, trd *tarReadData) error {
		err := trd.tarReadFileJSON(&trd.dockerManifestList)
		if err != nil {
			return err
		}
		trd.dockerManifestFound = true
		return nil
	}
}

// imageImportDockerAddLayerHandlers imports the docker layers when OCI import fails and docker manifest found.
func (rc *RegClient) imageImportDockerAddLayerHandlers(ctx context.Context, r ref.Ref, trd *tarReadData) {
	// remove handlers for OCI
	delete(trd.handlers, ociLayoutFilename)
	delete(trd.handlers, ociIndexFilename)

	index := 0
	if trd.name != "" {
		found := false
		tags := []string{}
		for i, entry := range trd.dockerManifestList {
			tags = append(tags, entry.RepoTags...)
			if slices.Contains(entry.RepoTags, trd.name) {
				index = i
				found = true
				break
			}
		}
		if !found {
			rc.slog.Warn("Could not find requested name",
				slog.Any("tags", tags),
				slog.String("name", trd.name))
			return
		}
	}

	// make a docker v2 manifest from first json array entry (can only tag one image)
	trd.dockerManifest.SchemaVersion = 2
	trd.dockerManifest.MediaType = mediatype.Docker2Manifest
	trd.dockerManifest.Layers = make([]descriptor.Descriptor, len(trd.dockerManifestList[index].Layers))

	// add handler for config
	trd.handlers[filepath.ToSlash(filepath.Clean(trd.dockerManifestList[index].Config))] = func(header *tar.Header, trd *tarReadData) error {
		// upload blob, digest is unknown
		d, err := rc.BlobPut(ctx, r, descriptor.Descriptor{Size: header.Size}, trd.tr)
		if err != nil {
			return err
		}
		// save the resulting descriptor to the manifest
		if od, ok := trd.dockerManifestList[index].LayerSources[d.Digest]; ok {
			trd.dockerManifest.Config = od
		} else {
			d.MediaType = mediatype.Docker2ImageConfig
			trd.dockerManifest.Config = d
		}
		return nil
	}
	// add handlers for each layer
	for i, layerFile := range trd.dockerManifestList[index].Layers {
		func(i int) {
			trd.handlers[filepath.ToSlash(filepath.Clean(layerFile))] = func(header *tar.Header, trd *tarReadData) error {
				// ensure blob is compressed
				rdrUC, err := archive.Decompress(trd.tr)
				if err != nil {
					return err
				}
				gzipR, err := archive.Compress(rdrUC, archive.CompressGzip)
				if err != nil {
					return err
				}
				defer gzipR.Close()
				// upload blob, digest and size is unknown
				d, err := rc.BlobPut(ctx, r, descriptor.Descriptor{}, gzipR)
				if err != nil {
					return err
				}
				// save the resulting descriptor in the appropriate layer
				if od, ok := trd.dockerManifestList[index].LayerSources[d.Digest]; ok {
					trd.dockerManifest.Layers[i] = od
				} else {
					d.MediaType = mediatype.Docker2LayerGzip
					trd.dockerManifest.Layers[i] = d
				}
				return nil
			}
		}(i)
	}
	trd.handleAdded = true
}

// imageImportOCIAddHandler adds handlers for oci-layout and index.json found in OCI layout tar files.
func (rc *RegClient) imageImportOCIAddHandler(ctx context.Context, r ref.Ref, trd *tarReadData) {
	// add handler for oci-layout, index.json, and manifest.json
	var err error
	var foundLayout, foundIndex bool

	// common handler code when both oci-layout and index.json have been processed
	ociHandler := func(trd *tarReadData) error {
		// no need to process docker manifest.json when OCI layout is available
		delete(trd.handlers, dockerManifestFilename)
		// create a manifest from the index
		trd.ociManifest, err = manifest.New(manifest.WithOrig(trd.ociIndex))
		if err != nil {
			return err
		}
		// start recursively processing manifests starting with the index
		// there's no need to push the index.json by digest, it will be pushed by tag if needed
		err = rc.imageImportOCIHandleManifest(ctx, r, trd.ociManifest, trd, false, false)
		if err != nil {
			return err
		}
		return nil
	}
	trd.handlers[ociLayoutFilename] = func(header *tar.Header, trd *tarReadData) error {
		var ociLayout v1.ImageLayout
		err := trd.tarReadFileJSON(&ociLayout)
		if err != nil {
			return err
		}
		if ociLayout.Version != ociLayoutVersion {
			// unknown version, ignore
			rc.slog.Warn("Unsupported oci-layout version",
				slog.String("version", ociLayout.Version))
			return nil
		}
		foundLayout = true
		if foundIndex {
			err = ociHandler(trd)
			if err != nil {
				return err
			}
		}
		return nil
	}
	trd.handlers[ociIndexFilename] = func(header *tar.Header, trd *tarReadData) error {
		err := trd.tarReadFileJSON(&trd.ociIndex)
		if err != nil {
			return err
		}
		foundIndex = true
		if foundLayout {
			err = ociHandler(trd)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// imageImportOCIHandleManifest recursively processes index and manifest entries from an OCI layout tar.
func (rc *RegClient) imageImportOCIHandleManifest(ctx context.Context, r ref.Ref, m manifest.Manifest, trd *tarReadData, push bool, child bool) error {
	// cache the manifest to avoid needing to pull again later, this is used if index.json is a wrapper around some other manifest
	trd.manifests[m.GetDescriptor().Digest] = m

	handleManifest := func(d descriptor.Descriptor, child bool) error {
		if err := d.Digest.Validate(); err != nil {
			return err
		}
		filename := tarOCILayoutDescPath(d)
		if !trd.processed[filename] && trd.handlers[filename] == nil {
			trd.handlers[filename] = func(header *tar.Header, trd *tarReadData) error {
				b, err := io.ReadAll(trd.tr)
				if err != nil {
					return err
				}
				switch d.MediaType {
				case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned,
					mediatype.Docker2Manifest, mediatype.Docker2ManifestList,
					mediatype.OCI1Manifest, mediatype.OCI1ManifestList:
					// known manifest media types
					md, err := manifest.New(manifest.WithDesc(d), manifest.WithRaw(b))
					if err != nil {
						return err
					}
					return rc.imageImportOCIHandleManifest(ctx, r, md, trd, true, child)
				case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
					mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
					mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
					mediatype.BuildkitCacheConfig:
					// known blob media types
					return rc.imageImportBlob(ctx, r, d, trd)
				default:
					// attempt manifest import, fall back to blob import
					md, err := manifest.New(manifest.WithDesc(d), manifest.WithRaw(b))
					if err == nil {
						return rc.imageImportOCIHandleManifest(ctx, r, md, trd, true, child)
					}
					return rc.imageImportBlob(ctx, r, d, trd)
				}
			}
		}
		return nil
	}

	if !push {
		mi, ok := m.(manifest.Indexer)
		if !ok {
			return fmt.Errorf("manifest doesn't support image methods%.0w", errs.ErrUnsupportedMediaType)
		}
		// for root index, add handler for matching reference (or only reference)
		dl, err := mi.GetManifestList()
		if err != nil {
			return err
		}
		// locate the digest in the index
		var d descriptor.Descriptor
		if len(dl) == 1 {
			d = dl[0]
		} else if r.Digest != "" {
			d.Digest = digest.Digest(r.Digest)
		} else if trd.name != "" {
			for _, cur := range dl {
				if cur.Annotations[annotationRefName] == trd.name {
					d = cur
					break
				}
			}
			if d.Digest.String() == "" {
				return fmt.Errorf("could not find requested tag in index.json, %s", trd.name)
			}
		} else {
			if r.Tag == "" {
				r.Tag = "latest"
			}
			// if more than one digest is in the index, use the first matching tag
			for _, cur := range dl {
				if cur.Annotations[annotationRefName] == r.Tag {
					d = cur
					break
				}
			}
			if d.Digest.String() == "" {
				return fmt.Errorf("could not find requested tag in index.json, %s", r.Tag)
			}
		}
		err = handleManifest(d, false)
		if err != nil {
			return err
		}
		// add a finish step to tag the selected digest
		trd.finish = append(trd.finish, func() error {
			mRef, ok := trd.manifests[d.Digest]
			if !ok {
				return fmt.Errorf("could not find manifest to tag, ref: %s, digest: %s", r.CommonName(), d.Digest)
			}
			return rc.ManifestPut(ctx, r, mRef)
		})
	} else if m.IsList() {
		// for index/manifest lists, add handlers for each embedded manifest
		mi, ok := m.(manifest.Indexer)
		if !ok {
			return fmt.Errorf("manifest doesn't support index methods%.0w", errs.ErrUnsupportedMediaType)
		}
		dl, err := mi.GetManifestList()
		if err != nil {
			return err
		}
		for _, d := range dl {
			err = handleManifest(d, true)
			if err != nil {
				return err
			}
		}
	} else {
		// else if a single image/manifest
		mi, ok := m.(manifest.Imager)
		if !ok {
			return fmt.Errorf("manifest doesn't support image methods%.0w", errs.ErrUnsupportedMediaType)
		}
		// add handler for the config descriptor if it's defined
		cd, err := mi.GetConfig()
		if err == nil {
			if err = cd.Digest.Validate(); err != nil {
				return err
			}
			filename := tarOCILayoutDescPath(cd)
			if !trd.processed[filename] && trd.handlers[filename] == nil {
				func(cd descriptor.Descriptor) {
					trd.handlers[filename] = func(header *tar.Header, trd *tarReadData) error {
						return rc.imageImportBlob(ctx, r, cd, trd)
					}
				}(cd)
			}
		}
		// add handlers for each layer
		layers, err := mi.GetLayers()
		if err != nil {
			return err
		}
		for _, d := range layers {
			if err = d.Digest.Validate(); err != nil {
				return err
			}
			filename := tarOCILayoutDescPath(d)
			if !trd.processed[filename] && trd.handlers[filename] == nil {
				func(d descriptor.Descriptor) {
					trd.handlers[filename] = func(header *tar.Header, trd *tarReadData) error {
						return rc.imageImportBlob(ctx, r, d, trd)
					}
				}(d)
			}
		}
	}
	// add a finish func to push the manifest, this gets skipped for the index.json
	if push {
		trd.finish = append(trd.finish, func() error {
			mRef := r.SetDigest(m.GetDescriptor().Digest.String())
			_, err := rc.ManifestHead(ctx, mRef)
			if err == nil {
				return nil
			}
			opts := []ManifestOpts{}
			if child {
				opts = append(opts, WithManifestChild())
			}
			return rc.ManifestPut(ctx, mRef, m, opts...)
		})
	}
	trd.handleAdded = true
	return nil
}

// imageImportOCIPushManifests uploads manifests after OCI blobs were successfully loaded.
func (rc *RegClient) imageImportOCIPushManifests(_ context.Context, _ ref.Ref, trd *tarReadData) error {
	// run finish handlers in reverse order to upload nested manifests
	for i := len(trd.finish) - 1; i >= 0; i-- {
		err := trd.finish[i]()
		if err != nil {
			return err
		}
	}
	return nil
}

func imagePlatformInList(target *platform.Platform, list []string) (bool, error) {
	// special case for an unset platform
	if target == nil || target.OS == "" {
		if slices.Contains(list, "") {
			return true, nil
		}
		return false, nil
	}
	for _, entry := range list {
		if entry == "" {
			continue
		}
		plat, err := platform.Parse(entry)
		if err != nil {
			return false, err
		}
		if platform.Match(*target, plat) {
			return true, nil
		}
	}
	return false, nil
}

// tarReadAll processes the tar file in a loop looking for matching filenames in the list of handlers.
// Handlers for filenames are added at the top level, and by manifest imports.
func (trd *tarReadData) tarReadAll(rs io.ReadSeeker) error {
	// return immediately if nothing to do
	if len(trd.handlers) == 0 {
		return nil
	}
	for {
		// reset back to beginning of tar file
		_, err := rs.Seek(0, 0)
		if err != nil {
			return err
		}
		dr, err := archive.Decompress(rs)
		if err != nil {
			return err
		}
		trd.tr = tar.NewReader(dr)
		trd.handleAdded = false
		// loop over each entry of the tar file
		for {
			header, err := trd.tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			name := filepath.ToSlash(filepath.Clean(header.Name))
			// track symlinks
			if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
				// normalize target relative to root of tar
				target := header.Linkname
				if !filepath.IsAbs(target) {
					target, err = filepath.Rel(filepath.Dir(name), target)
					if err != nil {
						return err
					}
				}
				target = filepath.ToSlash(filepath.Clean("/" + target)[1:])
				// track and set handleAdded if an existing handler points to the target
				if trd.linkAdd(name, target) && !trd.handleAdded {
					list, err := trd.linkList(target)
					if err != nil {
						return err
					}
					for _, src := range append(list, name) {
						if trd.handlers[src] != nil {
							trd.handleAdded = true
						}
					}
				}
			} else {
				// loop through filename and symlinks to file in search of handlers
				list, err := trd.linkList(name)
				if err != nil {
					return err
				}
				list = append(list, name)
				trdUsed := false
				for _, entry := range list {
					if trd.handlers[entry] != nil {
						// trd cannot be reused, force the loop to run again
						if trdUsed {
							trd.handleAdded = true
							break
						}
						trdUsed = true
						// run handler
						err = trd.handlers[entry](header, trd)
						if err != nil {
							return err
						}
						delete(trd.handlers, entry)
						trd.processed[entry] = true
						// return if last handler processed
						if len(trd.handlers) == 0 {
							return nil
						}
					}
				}
			}
		}
		// if entire file read without adding a new handler, fail
		if !trd.handleAdded {
			return fmt.Errorf("unable to read all files from tar: %w", errs.ErrNotFound)
		}
	}
}

func (trd *tarReadData) linkAdd(src, tgt string) bool {
	if slices.Contains(trd.links[tgt], src) {
		return false
	}
	trd.links[tgt] = append(trd.links[tgt], src)
	return true
}

func (trd *tarReadData) linkList(tgt string) ([]string, error) {
	list := trd.links[tgt]
	for _, entry := range list {
		if entry == tgt {
			return nil, fmt.Errorf("symlink loop encountered for %s", tgt)
		}
		list = append(list, trd.links[entry]...)
	}
	return list, nil
}

// tarReadFileJSON reads the current tar entry and unmarshals json into provided interface.
func (trd *tarReadData) tarReadFileJSON(data any) error {
	b, err := io.ReadAll(trd.tr)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, data)
	if err != nil {
		return err
	}
	return nil
}

var errTarFileExists = errors.New("tar file already exists")

func (td *tarWriteData) tarWriteHeader(filename string, size int64) error {
	dirName := filepath.ToSlash(filepath.Dir(filename))
	if !td.dirs[dirName] && dirName != "." {
		dirSplit := strings.Split(dirName, "/")
		for i := range dirSplit {
			dirJoin := strings.Join(dirSplit[:i+1], "/")
			if !td.dirs[dirJoin] && dirJoin != "" {
				header := tar.Header{
					Format:     tar.FormatPAX,
					Typeflag:   tar.TypeDir,
					Name:       dirJoin + "/",
					Size:       0,
					Mode:       td.mode | 0o511,
					ModTime:    td.timestamp,
					AccessTime: td.timestamp,
					ChangeTime: td.timestamp,
				}
				err := td.tw.WriteHeader(&header)
				if err != nil {
					return err
				}
				td.dirs[dirJoin] = true
			}
		}
	}
	if td.files[filename] {
		return fmt.Errorf("%w: %s", errTarFileExists, filename)
	}
	td.files[filename] = true
	header := tar.Header{
		Format:     tar.FormatPAX,
		Typeflag:   tar.TypeReg,
		Name:       filename,
		Size:       size,
		Mode:       td.mode | 0o400,
		ModTime:    td.timestamp,
		AccessTime: td.timestamp,
		ChangeTime: td.timestamp,
	}
	return td.tw.WriteHeader(&header)
}

func (td *tarWriteData) tarWriteFileJSON(filename string, data any) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = td.tarWriteHeader(filename, int64(len(dataJSON)))
	if err != nil {
		return err
	}
	_, err = td.tw.Write(dataJSON)
	if err != nil {
		return err
	}
	return nil
}

func tarOCILayoutDescPath(d descriptor.Descriptor) string {
	return fmt.Sprintf("blobs/%s/%s", d.Digest.Algorithm(), d.Digest.Encoded())
}
