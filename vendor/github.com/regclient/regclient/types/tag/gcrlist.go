// Contents in this file are from github.com/google/go-containerregistry

// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tag

import (
	"encoding/json"
	"strconv"
	"time"
)

type gcrRawManifestInfo struct {
	Size      string   `json:"imageSizeBytes"`
	MediaType string   `json:"mediaType"`
	Created   string   `json:"timeCreatedMs"`
	Uploaded  string   `json:"timeUploadedMs"`
	Tags      []string `json:"tag"`
}

// GCRManifestInfo is a Manifests entry is the output of List and Walk.
type GCRManifestInfo struct {
	Size      uint64    `json:"imageSizeBytes"`
	MediaType string    `json:"mediaType"`
	Created   time.Time `json:"timeCreatedMs"`
	Uploaded  time.Time `json:"timeUploadedMs"`
	Tags      []string  `json:"tag"`
}

func fromUnixMs(ms int64) time.Time {
	sec := ms / 1000
	ns := (ms % 1000) * 1000000
	return time.Unix(sec, ns)
}

func toUnixMs(t time.Time) string {
	return strconv.FormatInt(t.UnixNano()/1000000, 10)
}

// MarshalJSON implements json.Marshaler
func (m GCRManifestInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(gcrRawManifestInfo{
		Size:      strconv.FormatUint(m.Size, 10),
		MediaType: m.MediaType,
		Created:   toUnixMs(m.Created),
		Uploaded:  toUnixMs(m.Uploaded),
		Tags:      m.Tags,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (m *GCRManifestInfo) UnmarshalJSON(data []byte) error {
	raw := gcrRawManifestInfo{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Size != "" {
		size, err := strconv.ParseUint(raw.Size, 10, 64)
		if err != nil {
			return err
		}
		m.Size = size
	}

	if raw.Created != "" {
		created, err := strconv.ParseInt(raw.Created, 10, 64)
		if err != nil {
			return err
		}
		m.Created = fromUnixMs(created)
	}

	if raw.Uploaded != "" {
		uploaded, err := strconv.ParseInt(raw.Uploaded, 10, 64)
		if err != nil {
			return err
		}
		m.Uploaded = fromUnixMs(uploaded)
	}

	m.MediaType = raw.MediaType
	m.Tags = raw.Tags

	return nil
}
