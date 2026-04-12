package ocidir

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/ref"
)

// Close triggers a garbage collection if the underlying path has been modified
func (o *OCIDir) Close(ctx context.Context, r ref.Ref) error {
	if !o.gc {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if gc, ok := o.modRefs[r.Path]; !ok || !gc.mod || gc.locks > 0 {
		// unmodified or locked, skip gc
		return nil
	}

	// perform GC
	o.slog.Debug("running GC",
		slog.String("ref", r.CommonName()))
	dl := map[string]bool{}
	// recurse through index, manifests, and blob lists, generating a digest list
	index, err := o.readIndex(r, true)
	if err != nil {
		return err
	}
	im, err := manifest.New(manifest.WithOrig(index))
	if err != nil {
		return err
	}
	err = o.closeProcManifest(ctx, r, im, &dl)
	if err != nil {
		return err
	}

	// go through filesystem digest list, removing entries not seen in recursive pass
	blobsPath := path.Join(r.Path, "blobs")
	blobDirs, err := os.ReadDir(blobsPath)
	if err != nil {
		return err
	}
	for _, blobDir := range blobDirs {
		if !blobDir.IsDir() {
			// should this warn or delete unexpected files in the blobs folder?
			continue
		}
		digestFiles, err := os.ReadDir(path.Join(blobsPath, blobDir.Name()))
		if err != nil {
			return err
		}
		for _, digestFile := range digestFiles {
			digest := fmt.Sprintf("%s:%s", blobDir.Name(), digestFile.Name())
			if !dl[digest] {
				o.slog.Debug("ocidir garbage collect",
					slog.String("digest", digest))
				// delete
				err = os.Remove(path.Join(blobsPath, blobDir.Name(), digestFile.Name()))
				if err != nil {
					return fmt.Errorf("failed to delete %s: %w", path.Join(blobsPath, blobDir.Name(), digestFile.Name()), err)
				}
			}
		}
	}
	delete(o.modRefs, r.Path)
	return nil
}

func (o *OCIDir) closeProcManifest(ctx context.Context, r ref.Ref, m manifest.Manifest, dl *map[string]bool) error {
	if mi, ok := m.(manifest.Indexer); ok {
		// go through manifest list, updating dl, and recursively processing nested manifests
		ml, err := mi.GetManifestList()
		if err != nil {
			return err
		}
		for _, cur := range ml {
			cr := r.SetDigest(cur.Digest.String())
			(*dl)[cr.Digest] = true
			cm, err := o.manifestGet(ctx, cr)
			if err != nil {
				// ignore errors in case a manifest has been deleted or sparse copy
				o.slog.Debug("could not retrieve manifest",
					slog.String("ref", cr.CommonName()),
					slog.String("err", err.Error()))
				continue
			}
			err = o.closeProcManifest(ctx, cr, cm, dl)
			if err != nil {
				return err
			}
		}
	}
	if mi, ok := m.(manifest.Imager); ok {
		// get config from manifest if it exists
		cd, err := mi.GetConfig()
		if err == nil {
			(*dl)[cd.Digest.String()] = true
		}
		// finally add all layers to digest list
		layers, err := mi.GetLayers()
		if err != nil {
			return err
		}
		for _, layer := range layers {
			(*dl)[layer.Digest.String()] = true
		}
	}
	return nil
}
