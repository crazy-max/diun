// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"encoding/json"

	"go.mau.fi/util/exslices"
	"go.mau.fi/util/ptr"

	"maunium.net/go/mautrix/id"
)

type RelationType string

const (
	RelReplace             RelationType = "m.replace"
	RelReference           RelationType = "m.reference"
	RelAnnotation          RelationType = "m.annotation"
	RelThread              RelationType = "m.thread"
	RelBeeperTranscription RelationType = "com.beeper.transcription"
)

type RelatesTo struct {
	Type    RelationType `json:"rel_type,omitempty"`
	EventID id.EventID   `json:"event_id,omitempty"`
	Key     string       `json:"key,omitempty"`

	InReplyTo     *InReplyTo `json:"m.in_reply_to,omitempty"`
	IsFallingBack bool       `json:"is_falling_back,omitzero"`
}

type InReplyTo struct {
	EventID id.EventID `json:"event_id,omitempty"`

	UnstableRoomID id.RoomID       `json:"com.beeper.cross_room_id,omitempty"`
	BeeperQuote    json.RawMessage `json:"com.beeper.quote,omitempty"`
}

func (rel *RelatesTo) Copy() *RelatesTo {
	if rel == nil {
		return nil
	}
	clone := ptr.Clone(rel)
	clone.InReplyTo = ptr.Clone(clone.InReplyTo)
	return clone
}

func (rel *RelatesTo) GetReplaceID() id.EventID {
	if rel != nil && rel.Type == RelReplace {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetReferenceID() id.EventID {
	if rel != nil && rel.Type == RelReference {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetThreadParent() id.EventID {
	if rel != nil && rel.Type == RelThread {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetReplyTo() id.EventID {
	if rel != nil && rel.InReplyTo != nil {
		return rel.InReplyTo.EventID
	}
	return ""
}

func (rel *RelatesTo) GetNonFallbackReplyTo() id.EventID {
	if rel != nil && rel.InReplyTo != nil && (rel.Type != RelThread || !rel.IsFallingBack) {
		return rel.InReplyTo.EventID
	}
	return ""
}

func (rel *RelatesTo) GetAnnotationID() id.EventID {
	if rel != nil && rel.Type == RelAnnotation {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetAnnotationKey() string {
	if rel != nil && rel.Type == RelAnnotation {
		return rel.Key
	}
	return ""
}

func (rel *RelatesTo) SetReplace(mxid id.EventID) *RelatesTo {
	rel.Type = RelReplace
	rel.EventID = mxid
	return rel
}

func (rel *RelatesTo) SetReplyTo(mxid id.EventID) *RelatesTo {
	if rel.Type != RelThread {
		rel.Type = ""
		rel.EventID = ""
	}
	rel.InReplyTo = &InReplyTo{EventID: mxid}
	rel.IsFallingBack = false
	return rel
}

func (rel *RelatesTo) SetThread(mxid, fallback id.EventID) *RelatesTo {
	rel.Type = RelThread
	rel.EventID = mxid
	if fallback != "" && rel.GetReplyTo() == "" {
		rel.SetReplyTo(fallback)
		rel.IsFallingBack = true
	}
	return rel
}

func (rel *RelatesTo) SetAnnotation(mxid id.EventID, key string) *RelatesTo {
	rel.Type = RelAnnotation
	rel.EventID = mxid
	rel.Key = key
	return rel
}

type ReferenceAggregationItem struct {
	EventID id.EventID `json:"event_id"`
}

type ReferenceAggregation struct {
	Chunk []ReferenceAggregationItem `json:"chunk"`
}

func (ra *ReferenceAggregation) GetEventIDs() []id.EventID {
	if ra != nil && len(ra.Chunk) > 0 {
		return exslices.CastFunc(ra.Chunk, func(item ReferenceAggregationItem) id.EventID {
			return item.EventID
		})
	}
	return nil
}

type ThreadAggregation struct {
	LatestEvent             *Event `json:"latest_event"`
	Count                   int    `json:"count"`
	CurrentUserParticipated bool   `json:"current_user_participated"`
}

func (ta *ThreadAggregation) GetCount() int {
	if ta != nil {
		return ta.Count
	}
	return 0
}

func (ta *ThreadAggregation) GetCurrentUserParticipated() bool {
	if ta != nil {
		return ta.CurrentUserParticipated
	}
	return false
}

func (ta *ThreadAggregation) GetLatestEvent() *Event {
	if ta != nil {
		return ta.LatestEvent
	}
	return nil
}

type Relations struct {
	Thread     *ThreadAggregation    `json:"m.thread,omitempty"`
	References *ReferenceAggregation `json:"m.reference,omitempty"`
	LatestEdit *Event                `json:"m.replace,omitempty"`
}

func (rel *Relations) GetLatestEdit() *Event {
	if rel != nil {
		return rel.LatestEdit
	}
	return nil
}

func (rel *Relations) GetThread() *ThreadAggregation {
	if rel != nil {
		return rel.Thread
	}
	return nil
}

func (rel *Relations) GetReferences() *ReferenceAggregation {
	if rel != nil {
		return rel.References
	}
	return nil
}
