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

package ocidir

import (
	"context"
	"fmt"
	"os"

	"github.com/regclient/regclient/types/ping"
	"github.com/regclient/regclient/types/ref"
)

// Ping for an ocidir verifies access to read the path.
func (o *OCIDir) Ping(ctx context.Context, r ref.Ref) (ping.Result, error) {
	ret := ping.Result{}
	fd, err := os.Open(r.Path)
	if err != nil {
		return ret, err
	}
	defer fd.Close()
	fi, err := fd.Stat()
	if err != nil {
		return ret, err
	}
	ret.Stat = fi
	if !fi.IsDir() {
		return ret, fmt.Errorf("failed to access %s: not a directory", r.Path)
	}
	return ret, nil
}
