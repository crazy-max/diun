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

package schema2

import (
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/docker"
	"github.com/regclient/regclient/types/mediatype"
)

// ManifestListSchemaVersion is a pre-configured versioned field for manifest lists
var ManifestListSchemaVersion = docker.Versioned{
	SchemaVersion: 2,
	MediaType:     mediatype.Docker2ManifestList,
}

// ManifestList references manifests for various platforms.
type ManifestList struct {
	docker.Versioned

	// Manifests lists descriptors in the manifest list
	Manifests []descriptor.Descriptor `json:"manifests"`

	// Annotations contains arbitrary metadata for the image index.
	// Note, this is not a defined docker schema2 field.
	Annotations map[string]string `json:"annotations,omitempty"`
}
