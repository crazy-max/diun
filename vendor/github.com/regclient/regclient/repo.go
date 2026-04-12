package regclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/repo"
)

type repoLister interface {
	RepoList(ctx context.Context, hostname string, opts ...scheme.RepoOpts) (*repo.RepoList, error)
}

// RepoList returns a list of repositories on a registry.
// Note the underlying "_catalog" API is not supported on many cloud registries.
func (rc *RegClient) RepoList(ctx context.Context, hostname string, opts ...scheme.RepoOpts) (*repo.RepoList, error) {
	i := strings.Index(hostname, "/")
	if i > 0 {
		return nil, fmt.Errorf("invalid hostname: %s%.0w", hostname, errs.ErrParsingFailed)
	}
	schemeAPI, err := rc.schemeGet("reg")
	if err != nil {
		return nil, err
	}
	rl, ok := schemeAPI.(repoLister)
	if !ok {
		return nil, errs.ErrNotImplemented
	}
	return rl.RepoList(ctx, hostname, opts...)
}
