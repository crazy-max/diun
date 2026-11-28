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

package archive

import "errors"

var (
	// ErrNotImplemented used for routines that need to be developed still
	ErrNotImplemented = errors.New("this archive routine is not implemented yet")
	// ErrUnknownType used for unknown compression types
	ErrUnknownType = errors.New("unknown compression type")
	// ErrXzUnsupported because there isn't a Go package for this and I'm
	// avoiding dependencies on external binaries
	ErrXzUnsupported = errors.New("xz compression is currently unsupported")
)
