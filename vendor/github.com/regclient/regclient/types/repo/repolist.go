// Package repo handles a list of repositories from a registry
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/regclient/regclient/types/errs"
)

// RepoList is the response for a repository listing.
type RepoList struct {
	repoCommon
	RepoRegistryList
}

type repoCommon struct {
	host      string
	mt        string
	orig      any
	rawHeader http.Header
	rawBody   []byte
}

type repoConfig struct {
	host   string
	mt     string
	raw    []byte
	header http.Header
}

type Opts func(*repoConfig)

// New is used to create a repository listing.
func New(opts ...Opts) (*RepoList, error) {
	conf := repoConfig{
		mt: "application/json",
	}
	for _, opt := range opts {
		opt(&conf)
	}
	rl := RepoList{}
	rc := repoCommon{
		mt:        conf.mt,
		rawHeader: conf.header,
		rawBody:   conf.raw,
		host:      conf.host,
	}

	mt := strings.Split(conf.mt, ";")[0] // "application/json; charset=utf-8" -> "application/json"
	switch mt {
	case "application/json", "text/plain":
		err := json.Unmarshal(conf.raw, &rl)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: media type: %s, hostname: %s", errs.ErrUnsupportedMediaType, conf.mt, conf.host)
	}

	rl.repoCommon = rc
	return &rl, nil
}

func WithHeaders(header http.Header) Opts {
	return func(c *repoConfig) {
		c.header = header
	}
}

func WithHost(host string) Opts {
	return func(c *repoConfig) {
		c.host = host
	}
}

func WithMT(mt string) Opts {
	return func(c *repoConfig) {
		c.mt = mt
	}
}

func WithRaw(raw []byte) Opts {
	return func(c *repoConfig) {
		c.raw = raw
	}
}

// RepoRegistryList is a list of repositories from the _catalog API
type RepoRegistryList struct {
	Repositories []string `json:"repositories"`
}

func (r repoCommon) GetOrig() any {
	return r.orig
}

func (r repoCommon) MarshalJSON() ([]byte, error) {
	if len(r.rawBody) > 0 {
		return r.rawBody, nil
	}

	if r.orig != nil {
		return json.Marshal((r.orig))
	}
	return []byte{}, fmt.Errorf("JSON marshalling failed: %w", errs.ErrNotFound)
}

func (r repoCommon) RawBody() ([]byte, error) {
	return r.rawBody, nil
}

func (r repoCommon) RawHeaders() (http.Header, error) {
	return r.rawHeader, nil
}

// GetRepos returns the repositories
func (rl RepoRegistryList) GetRepos() ([]string, error) {
	return rl.Repositories, nil
}

// MarshalPretty is used for printPretty template formatting
func (rl RepoRegistryList) MarshalPretty() ([]byte, error) {
	sort.Slice(rl.Repositories, func(i, j int) bool {
		return strings.Compare(rl.Repositories[i], rl.Repositories[j]) < 0
	})
	buf := &bytes.Buffer{}
	for _, tag := range rl.Repositories {
		fmt.Fprintf(buf, "%s\n", tag)
	}
	return buf.Bytes(), nil
}
