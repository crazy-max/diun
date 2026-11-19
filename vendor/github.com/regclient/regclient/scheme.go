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

	"github.com/regclient/regclient/scheme"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
)

func (rc *RegClient) schemeGet(scheme string) (scheme.API, error) {
	s, ok := rc.schemes[scheme]
	if !ok {
		return nil, fmt.Errorf("%w: unknown scheme \"%s\"", errs.ErrNotImplemented, scheme)
	}
	return s, nil
}

// Close is used to free resources associated with a reference.
// With ocidir, this may trigger a garbage collection process.
func (rc *RegClient) Close(ctx context.Context, r ref.Ref) error {
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return err
	}
	// verify Closer api is defined, noop if missing
	sc, ok := schemeAPI.(scheme.Closer)
	if !ok {
		return nil
	}
	return sc.Close(ctx, r)
}
