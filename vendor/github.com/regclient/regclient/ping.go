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

	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
)

// Ping verifies access to a registry or equivalent.
func (rc *RegClient) Ping(ctx context.Context, r ref.Ref) (ping.Result, error) {
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return ping.Result{}, err
	}

	return schemeAPI.Ping(ctx, r)
}
