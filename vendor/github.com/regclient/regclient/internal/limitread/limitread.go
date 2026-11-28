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

// Package limitread provides a reader that will error if the limit is ever exceeded
package limitread

import (
	"fmt"
	"io"

	"github.com/regclient/regclient/types/errs"
)

type LimitRead struct {
	Reader io.Reader
	Limit  int64
}

func (lr *LimitRead) Read(p []byte) (int, error) {
	if lr.Limit < 0 {
		return 0, fmt.Errorf("read limit exceeded%.0w", errs.ErrSizeLimitExceeded)
	}
	if int64(len(p)) > lr.Limit+1 {
		p = p[0 : lr.Limit+1]
	}
	n, err := lr.Reader.Read(p)
	lr.Limit -= int64(n)
	if lr.Limit < 0 {
		return n, fmt.Errorf("read limit exceeded%.0w", errs.ErrSizeLimitExceeded)
	}
	return n, err
}
