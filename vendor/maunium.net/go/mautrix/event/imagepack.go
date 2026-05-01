// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"maunium.net/go/mautrix/id"
)

type ImagePackImage struct {
	URL  id.ContentURIString `json:"url"`
	Body string              `json:"body,omitempty"`
	Info *FileInfo           `json:"info,omitempty"`
}

type ImagePackUsage string

const (
	ImagePackUsageEmoji   ImagePackUsage = "emoticon"
	ImagePackUsageSticker ImagePackUsage = "sticker"
)

type ImagePackMetadata struct {
	DisplayName string              `json:"display_name,omitempty"`
	AvatarURL   id.ContentURIString `json:"avatar_url,omitempty"`
	Usage       []ImagePackUsage    `json:"usage,omitempty"`
	Attribution string              `json:"attribution,omitempty"`
}

func (ipm ImagePackMetadata) IsZero() bool {
	return ipm.DisplayName == "" && ipm.AvatarURL == "" && len(ipm.Usage) == 0 && ipm.Attribution == ""
}

type ImagePackEventContent struct {
	Images   map[string]*ImagePackImage `json:"images"`
	Metadata ImagePackMetadata          `json:"pack,omitzero"`
}

type ImagePackRoomsEventContent struct {
	Rooms map[id.RoomID]map[string]struct{} `json:"rooms"`
}
