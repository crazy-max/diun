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

package reg

import (
	"context"
	"fmt"

	"github.com/regclient/regclient/internal/reghttp"
	"github.com/regclient/regclient/internal/reqmeta"
	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
)

// Ping queries the /v2/ API of the registry to verify connectivity and access.
func (reg *Reg) Ping(ctx context.Context, r ref.Ref) (ping.Result, error) {
	ret := ping.Result{}
	req := &reghttp.Req{
		MetaKind:  reqmeta.Query,
		Host:      r.Registry,
		NoMirrors: true,
		Method:    "GET",
		Path:      "",
	}

	resp, err := reg.reghttp.Do(ctx, req)
	if resp != nil && resp.HTTPResponse() != nil {
		ret.Header = resp.HTTPResponse().Header
	}
	if err != nil {
		return ret, fmt.Errorf("failed to ping registry %s: %w", r.Registry, err)
	}
	defer resp.Close()

	if resp.HTTPResponse().StatusCode != 200 {
		return ret, fmt.Errorf("failed to ping registry %s: %w",
			r.Registry, reghttp.HTTPError(resp.HTTPResponse().StatusCode))
	}

	return ret, nil
}
