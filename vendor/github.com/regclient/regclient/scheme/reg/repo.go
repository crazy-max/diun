package reg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/regclient/regclient/internal/reghttp"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/repo"
)

// RepoList returns a list of repositories on a registry
// Note the underlying "_catalog" API is not supported on many cloud registries
func (reg *Reg) RepoList(ctx context.Context, hostname string, opts ...scheme.RepoOpts) (*repo.RepoList, error) {
	config := scheme.RepoConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	query := url.Values{}
	if config.Last != "" {
		query.Set("last", config.Last)
	}
	if config.Limit > 0 {
		query.Set("n", strconv.Itoa(config.Limit))
	}

	headers := http.Header{
		"Accept": []string{"application/json"},
	}
	req := &reghttp.Req{
		MetaKind:  reqmeta.Query,
		Host:      hostname,
		NoMirrors: true,
		Method:    "GET",
		Path:      "_catalog",
		NoPrefix:  true,
		Query:     query,
		Headers:   headers,
	}
	resp, err := reg.reghttp.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for %s: %w", hostname, err)
	}
	defer resp.Close()
	if resp.HTTPResponse().StatusCode != 200 {
		return nil, fmt.Errorf("failed to list repositories for %s: %w", hostname, reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	respBody, err := io.ReadAll(resp)
	if err != nil {
		reg.slog.Warn("Failed to read repo list",
			slog.String("err", err.Error()),
			slog.String("host", hostname))
		return nil, fmt.Errorf("failed to read repo list for %s: %w", hostname, err)
	}
	mt := mediatype.Base(resp.HTTPResponse().Header.Get("Content-Type"))
	rl, err := repo.New(
		repo.WithMT(mt),
		repo.WithRaw(respBody),
		repo.WithHost(hostname),
		repo.WithHeaders(resp.HTTPResponse().Header),
	)
	if err != nil {
		reg.slog.Warn("Failed to unmarshal repo list",
			slog.String("err", err.Error()),
			slog.String("body", string(respBody)),
			slog.String("host", hostname))
		return nil, fmt.Errorf("failed to parse repo list for %s: %w", hostname, err)
	}
	return rl, nil
}
