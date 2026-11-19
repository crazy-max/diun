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
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/referrer"
	"github.com/regclient/regclient/types/warning"
)

// ReferrerList retrieves a list of referrers to a manifest.
// The descriptor list should contain manifests that each have a subject field matching the requested ref.
func (rc *RegClient) ReferrerList(ctx context.Context, rSubject ref.Ref, opts ...scheme.ReferrerOpts) (referrer.ReferrerList, error) {
	if !rSubject.IsSet() {
		return referrer.ReferrerList{}, fmt.Errorf("ref is not set: %s%.0w", rSubject.CommonName(), errs.ErrInvalidReference)
	}
	// dedup warnings
	if w := warning.FromContext(ctx); w == nil {
		ctx = warning.NewContext(ctx, &warning.Warning{Hook: warning.DefaultHook()})
	}
	// set the digest on the subject reference
	config := scheme.ReferrerConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	if rSubject.Digest == "" || config.Platform != "" {
		mo := []ManifestOpts{WithManifestRequireDigest()}
		if config.Platform != "" {
			p, err := platform.Parse(config.Platform)
			if err != nil {
				return referrer.ReferrerList{}, fmt.Errorf("failed to lookup referrer platform: %w", err)
			}
			mo = append(mo, WithManifestPlatform(p))
		}
		m, err := rc.ManifestHead(ctx, rSubject, mo...)
		if err != nil {
			return referrer.ReferrerList{}, fmt.Errorf("failed to get digest for subject: %w", err)
		}
		rSubject = rSubject.SetDigest(m.GetDescriptor().Digest.String())
	}
	// lookup the scheme for the appropriate ref
	var r ref.Ref
	if config.SrcRepo.IsSet() {
		r = config.SrcRepo
	} else {
		r = rSubject
	}
	schemeAPI, err := rc.schemeGet(r.Scheme)
	if err != nil {
		return referrer.ReferrerList{}, err
	}
	return schemeAPI.ReferrerList(ctx, rSubject, opts...)
}
