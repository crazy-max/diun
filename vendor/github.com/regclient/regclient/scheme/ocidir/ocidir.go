// Package ocidir implements the OCI Image Layout scheme with a directory (not packed in a tar)
package ocidir

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/regclient/regclient/internal/pqueue"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

const (
	imageLayoutFile = "oci-layout"
	aOCIRefName     = "org.opencontainers.image.ref.name"
	aCtrdImageName  = "io.containerd.image.name"
	defThrottle     = 3
)

// OCIDir is used for accessing OCI Image Layouts defined as a directory
type OCIDir struct {
	slog        *slog.Logger
	gc          bool
	modRefs     map[string]*ociGC
	throttle    map[string]*pqueue.Queue[reqmeta.Data]
	throttleDef int
	mu          sync.Mutex
}

type ociGC struct {
	mod   bool
	locks int
}

type ociConf struct {
	gc       bool
	slog     *slog.Logger
	throttle int
}

// Opts are used for passing options to ocidir
type Opts func(*ociConf)

// New creates a new OCIDir with options
func New(opts ...Opts) *OCIDir {
	conf := ociConf{
		slog:     slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		gc:       true,
		throttle: defThrottle,
	}
	for _, opt := range opts {
		opt(&conf)
	}
	return &OCIDir{
		slog:        conf.slog,
		gc:          conf.gc,
		modRefs:     map[string]*ociGC{},
		throttle:    map[string]*pqueue.Queue[reqmeta.Data]{},
		throttleDef: conf.throttle,
	}
}

// WithGC configures the garbage collection setting
// This defaults to enabled
func WithGC(gc bool) Opts {
	return func(c *ociConf) {
		c.gc = gc
	}
}

// WithSlog provides a slog logger.
// By default logging is disabled.
func WithSlog(slog *slog.Logger) Opts {
	return func(c *ociConf) {
		c.slog = slog
	}
}

// WithThrottle provides a number of concurrent write actions (blob/manifest put)
func WithThrottle(count int) Opts {
	return func(c *ociConf) {
		c.throttle = count
	}
}

// GCLock is used to prevent GC on a ref
func (o *OCIDir) GCLock(r ref.Ref) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if gc, ok := o.modRefs[r.Path]; ok && gc != nil {
		gc.locks++
	} else {
		o.modRefs[r.Path] = &ociGC{locks: 1}
	}
}

// GCUnlock removes a hold on GC of a ref, this must be done before the ref is closed
func (o *OCIDir) GCUnlock(r ref.Ref) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if gc, ok := o.modRefs[r.Path]; ok && gc != nil && gc.locks > 0 {
		gc.locks--
	}
}

// Throttle is used to limit concurrency
func (o *OCIDir) Throttle(r ref.Ref, put bool) []*pqueue.Queue[reqmeta.Data] {
	tList := []*pqueue.Queue[reqmeta.Data]{}
	// throttle only applies to put requests
	if !put || o.throttleDef <= 0 {
		return tList
	}
	return []*pqueue.Queue[reqmeta.Data]{o.throttleGet(r, false)}
}

func (o *OCIDir) throttleGet(r ref.Ref, locked bool) *pqueue.Queue[reqmeta.Data] {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	if t, ok := o.throttle[r.Path]; ok {
		return t
	}
	// init a new throttle
	o.throttle[r.Path] = pqueue.New(pqueue.Opts[reqmeta.Data]{Max: o.throttleDef})
	return o.throttle[r.Path]
}

func (o *OCIDir) initIndex(r ref.Ref, locked bool) error {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	layoutFile := path.Join(r.Path, imageLayoutFile)
	_, err := os.Stat(layoutFile)
	if err == nil {
		return nil
	}
	//#nosec G301 defer to user umask settings
	err = os.MkdirAll(r.Path, 0o777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("failed creating %s: %w", r.Path, err)
	}
	// create/replace oci-layout file
	layout := v1.ImageLayout{
		Version: "1.0.0",
	}
	lb, err := json.Marshal(layout)
	if err != nil {
		return fmt.Errorf("cannot marshal layout: %w", err)
	}
	//#nosec G304 users should validate references they attempt to open
	lfh, err := os.Create(layoutFile)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", imageLayoutFile, err)
	}
	defer lfh.Close()
	_, err = lfh.Write(lb)
	if err != nil {
		return fmt.Errorf("cannot write %s: %w", imageLayoutFile, err)
	}
	return nil
}

func (o *OCIDir) readIndex(r ref.Ref, locked bool) (v1.Index, error) {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	// validate dir
	index := v1.Index{}
	err := o.valid(r.Path, true)
	if err != nil {
		return index, err
	}
	indexFile := path.Join(r.Path, "index.json")
	//#nosec G304 users should validate references they attempt to open
	fh, err := os.Open(indexFile)
	if err != nil {
		return index, fmt.Errorf("%s cannot be open: %w", indexFile, err)
	}
	defer fh.Close()
	ib, err := io.ReadAll(fh)
	if err != nil {
		return index, fmt.Errorf("%s cannot be read: %w", indexFile, err)
	}
	err = json.Unmarshal(ib, &index)
	if err != nil {
		return index, fmt.Errorf("%s cannot be parsed: %w", indexFile, err)
	}
	return index, nil
}

func (o *OCIDir) updateIndex(r ref.Ref, d descriptor.Descriptor, child bool, locked bool) error {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	indexChanged := false
	index, err := o.readIndex(r, true)
	if err != nil {
		index = indexCreate()
		indexChanged = true
	}
	if !child {
		err := indexSet(&index, r, d)
		if err != nil {
			return fmt.Errorf("failed to update index: %w", err)
		}
		indexChanged = true
	}
	if indexChanged {
		err = o.writeIndex(r, index, true)
		if err != nil {
			return fmt.Errorf("failed to write index: %w", err)
		}
	}
	return nil
}

func (o *OCIDir) writeIndex(r ref.Ref, i v1.Index, locked bool) error {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	//#nosec G301 defer to user umask settings
	err := os.MkdirAll(r.Path, 0o777)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("failed creating %s: %w", r.Path, err)
	}
	// create/replace oci-layout file
	layout := v1.ImageLayout{
		Version: "1.0.0",
	}
	lb, err := json.Marshal(layout)
	if err != nil {
		return fmt.Errorf("cannot marshal layout: %w", err)
	}
	lfh, err := os.Create(path.Join(r.Path, imageLayoutFile))
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", imageLayoutFile, err)
	}
	defer lfh.Close()
	_, err = lfh.Write(lb)
	if err != nil {
		return fmt.Errorf("cannot write %s: %w", imageLayoutFile, err)
	}
	// create/replace index.json file
	tmpFile, err := os.CreateTemp(r.Path, "index.json.*.tmp")
	if err != nil {
		return fmt.Errorf("cannot create index tmpfile: %w", err)
	}
	fi, err := tmpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat index tmpfile: %w", err)
	}
	tmpName := fi.Name()
	b, err := json.Marshal(i)
	if err != nil {
		return fmt.Errorf("cannot marshal index: %w", err)
	}
	_, err = tmpFile.Write(b)
	errC := tmpFile.Close()
	if err != nil {
		return fmt.Errorf("cannot write index: %w", err)
	}
	if errC != nil {
		return fmt.Errorf("cannot close index: %w", errC)
	}
	indexFile := path.Join(r.Path, "index.json")
	//#nosec G703 inputs are user controlled
	err = os.Rename(path.Join(r.Path, tmpName), indexFile)
	if err != nil {
		return fmt.Errorf("cannot rename tmpfile to index: %w", err)
	}
	return nil
}

// func valid (dir) (error) // check for `oci-layout` file and `index.json` for read
func (o *OCIDir) valid(dir string, locked bool) error {
	if !locked {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	layout := v1.ImageLayout{}
	reqVer := "1.0.0"
	//#nosec G304 users should validate references they attempt to open
	fh, err := os.Open(path.Join(dir, imageLayoutFile))
	if err != nil {
		return fmt.Errorf("%s cannot be open: %w", imageLayoutFile, err)
	}
	defer fh.Close()
	lb, err := io.ReadAll(fh)
	if err != nil {
		return fmt.Errorf("%s cannot be read: %w", imageLayoutFile, err)
	}
	err = json.Unmarshal(lb, &layout)
	if err != nil {
		return fmt.Errorf("%s cannot be parsed: %w", imageLayoutFile, err)
	}
	if layout.Version != reqVer {
		return fmt.Errorf("unsupported oci layout version, expected %s, received %s", reqVer, layout.Version)
	}
	return nil
}

func (o *OCIDir) refMod(r ref.Ref) {
	if gc, ok := o.modRefs[r.Path]; ok && gc != nil {
		gc.mod = true
	} else {
		o.modRefs[r.Path] = &ociGC{mod: true}
	}
}

func indexCreate() v1.Index {
	i := v1.Index{
		Versioned:   v1.IndexSchemaVersion,
		MediaType:   mediatype.OCI1ManifestList,
		Manifests:   []descriptor.Descriptor{},
		Annotations: map[string]string{},
	}
	return i
}

func indexGet(index v1.Index, r ref.Ref) (descriptor.Descriptor, error) {
	if r.Digest == "" && r.Tag == "" {
		r = r.SetTag("latest")
	}
	if r.Digest != "" {
		for _, im := range index.Manifests {
			if im.Digest.String() == r.Digest {
				return im, nil
			}
		}
	} else if r.Tag != "" {
		for _, im := range index.Manifests {
			if name, ok := im.Annotations[aOCIRefName]; ok && name == r.Tag {
				return im, nil
			}
		}
		// fall back to support full image name in annotation
		for _, im := range index.Manifests {
			if name, ok := im.Annotations[aOCIRefName]; ok && strings.HasSuffix(name, ":"+r.Tag) {
				return im, nil
			}
		}
	}
	return descriptor.Descriptor{}, errs.ErrNotFound
}

func indexSet(index *v1.Index, r ref.Ref, d descriptor.Descriptor) error {
	if index == nil {
		return fmt.Errorf("index is nil")
	}
	if r.Tag != "" {
		if d.Annotations == nil {
			d.Annotations = map[string]string{}
		}
		d.Annotations[aOCIRefName] = r.Tag
	}
	if index.Manifests == nil {
		index.Manifests = []descriptor.Descriptor{}
	}
	pos := -1
	// search for existing
	for i := range index.Manifests {
		var name string
		if index.Manifests[i].Annotations != nil {
			name = index.Manifests[i].Annotations[aOCIRefName]
		}
		if (name == "" && index.Manifests[i].Digest == d.Digest) || (r.Tag != "" && name == r.Tag) {
			index.Manifests[i] = d
			pos = i
			break
		}
	}
	if pos >= 0 {
		// existing entry was replaced, remove any dup entries
		for i := len(index.Manifests) - 1; i > pos; i-- {
			var name string
			if index.Manifests[i].Annotations != nil {
				name = index.Manifests[i].Annotations[aOCIRefName]
			}
			// prune entries without any tag and a matching digest
			// or entries with a matching tag
			if (name == "" && index.Manifests[i].Digest == d.Digest) || (r.Tag != "" && name == r.Tag) {
				index.Manifests = slices.Delete(index.Manifests, i, i+1)
			}
		}
	} else {
		// existing entry to replace was not found, add the descriptor
		index.Manifests = append(index.Manifests, d)
	}
	return nil
}
