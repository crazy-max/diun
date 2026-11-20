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

// Package ping is used for data types with the Ping methods.
package ping

import (
	"io/fs"
	"net/http"
)

// Result is the response to a ping request.
type Result struct {
	Header http.Header // Header is defined for responses from a registry.
	Stat   fs.FileInfo // Stat is defined for responses from an ocidir.
}
