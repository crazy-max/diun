// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
