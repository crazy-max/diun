package tag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/mediatype"
	ociv1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

// List contains a tag list.
// Currently this is a struct but the underlying type could be changed to an interface in the future.
// Using methods is recommended over directly accessing fields.
type List struct {
	tagCommon
	DockerList
	GCRList
	LayoutList
}

type tagCommon struct {
	r         ref.Ref
	mt        string
	orig      any
	rawHeader http.Header
	rawBody   []byte
	url       *url.URL
}

// DockerList is returned from registry/2.0 API's.
type DockerList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GCRList fields are from gcr.io.
type GCRList struct {
	Children  []string                   `json:"child,omitempty"`
	Manifests map[string]GCRManifestInfo `json:"manifest,omitempty"`
}

// LayoutList includes the OCI Index from an OCI Layout.
type LayoutList struct {
	Index ociv1.Index
}

type tagConfig struct {
	ref    ref.Ref
	mt     string
	raw    []byte
	header http.Header
	index  ociv1.Index
	tags   []string
	url    *url.URL
}

// Opts defines options for creating a new tag.
type Opts func(*tagConfig)

// New creates a tag list from options.
// Tags may be provided directly, or they will be parsed from the raw input based on the media type.
func New(opts ...Opts) (*List, error) {
	conf := tagConfig{}
	for _, opt := range opts {
		opt(&conf)
	}
	if conf.mt == "" {
		conf.mt = "application/json"
	}
	tl := List{}
	tc := tagCommon{
		r:         conf.ref,
		mt:        conf.mt,
		rawHeader: conf.header,
		rawBody:   conf.raw,
		url:       conf.url,
	}
	if len(conf.tags) > 0 {
		tl.Tags = conf.tags
	}
	if conf.index.Manifests != nil {
		tl.LayoutList.Index = conf.index
	}
	if len(conf.raw) > 0 {
		mt := mediatype.Base(conf.mt)
		switch mt {
		case "application/json", "text/plain":
			err := json.Unmarshal(conf.raw, &tl)
			if err != nil {
				return nil, err
			}
		case mediatype.OCI1ManifestList:
			// noop
		default:
			return nil, fmt.Errorf("%w: media type: %s, reference: %s", errs.ErrUnsupportedMediaType, conf.mt, conf.ref.CommonName())
		}
	}
	tl.tagCommon = tc

	return &tl, nil
}

// WithHeaders includes data from http headers when creating tag list.
func WithHeaders(header http.Header) Opts {
	return func(tConf *tagConfig) {
		tConf.header = header
	}
}

// WithLayoutIndex include the index from an OCI Layout.
func WithLayoutIndex(index ociv1.Index) Opts {
	return func(tConf *tagConfig) {
		tConf.index = index
	}
}

// WithMT sets the returned media type on the tag list.
func WithMT(mt string) Opts {
	return func(tConf *tagConfig) {
		tConf.mt = mt
	}
}

// WithRaw defines the raw response from the tag list request.
func WithRaw(raw []byte) Opts {
	return func(tConf *tagConfig) {
		tConf.raw = raw
	}
}

// WithRef specifies the reference (repository) associated with the tag list.
func WithRef(ref ref.Ref) Opts {
	return func(tConf *tagConfig) {
		tConf.ref = ref
	}
}

// WithResp includes the response from an http request.
func WithResp(resp *http.Response) Opts {
	return func(tConf *tagConfig) {
		if len(tConf.raw) == 0 {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				tConf.raw = body
			}
		}
		if tConf.header == nil {
			tConf.header = resp.Header
		}
		if tConf.mt == "" && resp.Header != nil {
			tConf.mt = resp.Header.Get("Content-Type")
		}
		if tConf.url == nil {
			tConf.url = resp.Request.URL
		}
	}
}

// WithTags provides the parsed tags for the tag list.
func WithTags(tags []string) Opts {
	return func(tConf *tagConfig) {
		tConf.tags = tags
	}
}

// Append extends a tag list with another.
func (l *List) Append(add *List) error {
	// verify two lists are compatible
	if l.mt != add.mt || !ref.EqualRepository(l.r, add.r) || l.Name != add.Name {
		return fmt.Errorf("unable to append, lists are incompatible")
	}
	if add.orig != nil {
		l.orig = add.orig
	}
	if add.rawBody != nil {
		l.rawBody = add.rawBody
	}
	if add.rawHeader != nil {
		l.rawHeader = add.rawHeader
	}
	if add.url != nil {
		l.url = add.url
	}
	l.Tags = append(l.Tags, add.Tags...)
	if add.Children != nil {
		l.Children = append(l.Children, add.Children...)
	}
	if add.Manifests != nil {
		if l.Manifests == nil {
			l.Manifests = add.Manifests
		} else {
			maps.Copy(l.Manifests, add.Manifests)
		}
	}
	return nil
}

// GetOrig returns the underlying tag data structure if defined.
func (t tagCommon) GetOrig() any {
	return t.orig
}

// MarshalJSON returns the tag list in json.
func (t tagCommon) MarshalJSON() ([]byte, error) {
	if len(t.rawBody) > 0 {
		return t.rawBody, nil
	}

	if t.orig != nil {
		return json.Marshal((t.orig))
	}
	return []byte{}, fmt.Errorf("JSON marshalling failed: %w", errs.ErrNotFound)
}

// RawBody returns the original tag list response.
func (t tagCommon) RawBody() ([]byte, error) {
	return t.rawBody, nil
}

// RawHeaders returns the received http headers.
func (t tagCommon) RawHeaders() (http.Header, error) {
	return t.rawHeader, nil
}

// GetURL returns the URL of the request.
func (t tagCommon) GetURL() *url.URL {
	return t.url
}

// GetTags returns the tags from a list.
func (tl DockerList) GetTags() ([]string, error) {
	return tl.Tags, nil
}

// MarshalPretty is used for printPretty template formatting.
func (tl DockerList) MarshalPretty() ([]byte, error) {
	sort.Slice(tl.Tags, func(i, j int) bool {
		return strings.Compare(tl.Tags[i], tl.Tags[j]) < 0
	})
	buf := &bytes.Buffer{}
	for _, tag := range tl.Tags {
		fmt.Fprintf(buf, "%s\n", tag)
	}
	return buf.Bytes(), nil
}
