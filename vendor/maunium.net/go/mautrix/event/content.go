// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
)

// TypeMap is a mapping from event type to the content struct type.
// This is used by Content.ParseRaw() for creating the correct type of struct.
var TypeMap = map[Type]reflect.Type{
	StateMember:            reflect.TypeFor[MemberEventContent](),
	StateThirdPartyInvite:  reflect.TypeFor[ThirdPartyInviteEventContent](),
	StatePowerLevels:       reflect.TypeFor[PowerLevelsEventContent](),
	StateCanonicalAlias:    reflect.TypeFor[CanonicalAliasEventContent](),
	StateRoomName:          reflect.TypeFor[RoomNameEventContent](),
	StateRoomAvatar:        reflect.TypeFor[RoomAvatarEventContent](),
	StateServerACL:         reflect.TypeFor[ServerACLEventContent](),
	StateTopic:             reflect.TypeFor[TopicEventContent](),
	StateTombstone:         reflect.TypeFor[TombstoneEventContent](),
	StateCreate:            reflect.TypeFor[CreateEventContent](),
	StateJoinRules:         reflect.TypeFor[JoinRulesEventContent](),
	StateHistoryVisibility: reflect.TypeFor[HistoryVisibilityEventContent](),
	StateGuestAccess:       reflect.TypeFor[GuestAccessEventContent](),
	StatePinnedEvents:      reflect.TypeFor[PinnedEventsEventContent](),
	StatePolicyRoom:        reflect.TypeFor[ModPolicyContent](),
	StatePolicyServer:      reflect.TypeFor[ModPolicyContent](),
	StatePolicyUser:        reflect.TypeFor[ModPolicyContent](),
	StateEncryption:        reflect.TypeFor[EncryptionEventContent](),
	StateBridge:            reflect.TypeFor[BridgeEventContent](),
	StateHalfShotBridge:    reflect.TypeFor[BridgeEventContent](),
	StateSpaceParent:       reflect.TypeFor[SpaceParentEventContent](),
	StateSpaceChild:        reflect.TypeFor[SpaceChildEventContent](),

	StateRoomPolicy: reflect.TypeFor[RoomPolicyEventContent](),

	StateImagePack:         reflect.TypeFor[ImagePackEventContent](),
	StateUnstableImagePack: reflect.TypeFor[ImagePackEventContent](),

	StateLegacyPolicyRoom:     reflect.TypeFor[ModPolicyContent](),
	StateLegacyPolicyServer:   reflect.TypeFor[ModPolicyContent](),
	StateLegacyPolicyUser:     reflect.TypeFor[ModPolicyContent](),
	StateUnstablePolicyRoom:   reflect.TypeFor[ModPolicyContent](),
	StateUnstablePolicyServer: reflect.TypeFor[ModPolicyContent](),
	StateUnstablePolicyUser:   reflect.TypeFor[ModPolicyContent](),

	StateElementFunctionalMembers: reflect.TypeFor[ElementFunctionalMembersContent](),
	StateBeeperRoomFeatures:       reflect.TypeFor[RoomFeatures](),
	StateBeeperDisappearingTimer:  reflect.TypeFor[BeeperDisappearingTimer](),

	EventMessage:   reflect.TypeFor[MessageEventContent](),
	EventSticker:   reflect.TypeFor[MessageEventContent](),
	EventEncrypted: reflect.TypeFor[EncryptedEventContent](),
	EventRedaction: reflect.TypeFor[RedactionEventContent](),
	EventReaction:  reflect.TypeFor[ReactionEventContent](),

	EventUnstablePollStart:    reflect.TypeFor[PollStartEventContent](),
	EventUnstablePollResponse: reflect.TypeFor[PollResponseEventContent](),

	BeeperMessageStatus:          reflect.TypeFor[BeeperMessageStatusEventContent](),
	BeeperTranscription:          reflect.TypeFor[BeeperTranscriptionEventContent](),
	BeeperViewLimitedMediaUpdate: reflect.TypeFor[BeeperViewLimitedMediaUpdateContent](),
	BeeperDeleteChat:             reflect.TypeFor[BeeperChatDeleteEventContent](),
	BeeperBlockUser:              reflect.TypeFor[BeeperBlockUserEventContent](),
	BeeperAcceptMessageRequest:   reflect.TypeFor[BeeperAcceptMessageRequestEventContent](),
	BeeperSendState:              reflect.TypeFor[BeeperSendStateEventContent](),

	AccountDataRoomTags:           reflect.TypeFor[TagEventContent](),
	AccountDataDirectChats:        reflect.TypeFor[DirectChatsEventContent](),
	AccountDataFullyRead:          reflect.TypeFor[FullyReadEventContent](),
	AccountDataIgnoredUserList:    reflect.TypeFor[IgnoredUserListEventContent](),
	AccountDataMarkedUnread:       reflect.TypeFor[MarkedUnreadEventContent](),
	AccountDataBeeperMute:         reflect.TypeFor[BeeperMuteEventContent](),
	AccountDataPerMessageProfiles: reflect.TypeFor[PerMessageProfilesEventContent](),

	AccountDataImagePackRooms:         reflect.TypeFor[ImagePackRoomsEventContent](),
	AccountDataUnstableImagePackRooms: reflect.TypeFor[ImagePackRoomsEventContent](),

	EphemeralEventTyping:   reflect.TypeFor[TypingEventContent](),
	EphemeralEventReceipt:  reflect.TypeFor[ReceiptEventContent](),
	EphemeralEventPresence: reflect.TypeFor[PresenceEventContent](),

	InRoomVerificationReady:  reflect.TypeFor[VerificationReadyEventContent](),
	InRoomVerificationStart:  reflect.TypeFor[VerificationStartEventContent](),
	InRoomVerificationDone:   reflect.TypeFor[VerificationDoneEventContent](),
	InRoomVerificationCancel: reflect.TypeFor[VerificationCancelEventContent](),

	InRoomVerificationAccept: reflect.TypeFor[VerificationAcceptEventContent](),
	InRoomVerificationKey:    reflect.TypeFor[VerificationKeyEventContent](),
	InRoomVerificationMAC:    reflect.TypeFor[VerificationMACEventContent](),

	ToDeviceRoomKey:          reflect.TypeFor[RoomKeyEventContent](),
	ToDeviceForwardedRoomKey: reflect.TypeFor[ForwardedRoomKeyEventContent](),
	ToDeviceRoomKeyBundle:    reflect.TypeFor[RoomKeyBundleEventContent](),
	ToDeviceRoomKeyRequest:   reflect.TypeFor[RoomKeyRequestEventContent](),
	ToDeviceEncrypted:        reflect.TypeFor[EncryptedEventContent](),
	ToDeviceRoomKeyWithheld:  reflect.TypeFor[RoomKeyWithheldEventContent](),
	ToDeviceSecretRequest:    reflect.TypeFor[SecretRequestEventContent](),
	ToDeviceSecretSend:       reflect.TypeFor[SecretSendEventContent](),
	ToDeviceSecretPush:       reflect.TypeFor[SecretPushEventContent](),
	ToDeviceDummy:            reflect.TypeFor[DummyEventContent](),

	ToDeviceVerificationRequest: reflect.TypeFor[VerificationRequestEventContent](),
	ToDeviceVerificationReady:   reflect.TypeFor[VerificationReadyEventContent](),
	ToDeviceVerificationStart:   reflect.TypeFor[VerificationStartEventContent](),
	ToDeviceVerificationDone:    reflect.TypeFor[VerificationDoneEventContent](),
	ToDeviceVerificationCancel:  reflect.TypeFor[VerificationCancelEventContent](),

	ToDeviceVerificationAccept: reflect.TypeFor[VerificationAcceptEventContent](),
	ToDeviceVerificationKey:    reflect.TypeFor[VerificationKeyEventContent](),
	ToDeviceVerificationMAC:    reflect.TypeFor[VerificationMACEventContent](),

	ToDeviceOrgMatrixRoomKeyWithheld: reflect.TypeFor[RoomKeyWithheldEventContent](),

	ToDeviceBeeperRoomKeyAck:      reflect.TypeFor[BeeperRoomKeyAckEventContent](),
	ToDeviceBeeperStreamSubscribe: reflect.TypeFor[BeeperStreamSubscribeEventContent](),
	ToDeviceBeeperStreamUpdate:    reflect.TypeFor[BeeperStreamUpdateEventContent](),

	CallInvite:       reflect.TypeFor[CallInviteEventContent](),
	CallCandidates:   reflect.TypeFor[CallCandidatesEventContent](),
	CallAnswer:       reflect.TypeFor[CallAnswerEventContent](),
	CallReject:       reflect.TypeFor[CallRejectEventContent](),
	CallSelectAnswer: reflect.TypeFor[CallSelectAnswerEventContent](),
	CallNegotiate:    reflect.TypeFor[CallNegotiateEventContent](),
	CallHangup:       reflect.TypeFor[CallHangupEventContent](),
}

// Content stores the content of a Matrix event.
//
// By default, the raw JSON bytes are stored in VeryRaw and parsed into a map[string]interface{} in the Raw field.
// Additionally, you can call ParseRaw with the correct event type to parse the (VeryRaw) content into a nicer struct,
// which you can then access from Parsed or via the helper functions.
//
// When being marshaled into JSON, the data in Parsed will be marshaled first and then recursively merged
// with the data in Raw. Values in Raw are preferred, but nested objects will be recursed into before merging,
// rather than overriding the whole object with the one in Raw).
// If one of them is nil, then only the other is used. If both (Parsed and Raw) are nil, VeryRaw is used instead.
type Content struct {
	VeryRaw json.RawMessage
	Raw     map[string]any
	Parsed  any
}

type Relatable interface {
	GetRelatesTo() *RelatesTo
	OptionalGetRelatesTo() *RelatesTo
	SetRelatesTo(rel *RelatesTo)
}

func (content *Content) UnmarshalJSON(data []byte) error {
	content.VeryRaw = bytes.Clone(data)
	err := json.Unmarshal(data, &content.Raw)
	return err
}

func (content *Content) MarshalJSON() ([]byte, error) {
	if content.Raw == nil {
		if content.Parsed == nil {
			if content.VeryRaw == nil {
				return []byte("{}"), nil
			}
			return content.VeryRaw, nil
		}
		return json.Marshal(content.Parsed)
	} else if content.Parsed != nil {
		// TODO this whole thing is incredibly hacky
		// It needs to produce JSON, where:
		// * content.Parsed is applied after content.Raw
		// * MarshalJSON() is respected inside content.Parsed
		// * Custom field inside nested objects of content.Raw are preserved,
		//   even if content.Parsed contains the higher-level objects.
		// * content.Raw is not modified

		unparsed, err := json.Marshal(content.Parsed)
		if err != nil {
			return nil, err
		}

		var rawParsed map[string]any
		err = json.Unmarshal(unparsed, &rawParsed)
		if err != nil {
			return nil, err
		}

		output := make(map[string]any)
		maps.Copy(output, content.Raw)

		mergeMaps(output, rawParsed)
		return json.Marshal(output)
	}
	return json.Marshal(content.Raw)
}

// Deprecated: use errors.Is directly
func IsUnsupportedContentType(err error) bool {
	return errors.Is(err, ErrUnsupportedContentType)
}

var ErrContentAlreadyParsed = errors.New("content is already parsed")
var ErrUnsupportedContentType = errors.New("unsupported event type")

func (content *Content) GetRaw() map[string]any {
	if content.Raw == nil {
		content.Raw = make(map[string]any)
	}
	return content.Raw
}

func (content *Content) ParseRaw(evtType Type) error {
	if content.Parsed != nil {
		return ErrContentAlreadyParsed
	}
	structType, ok := TypeMap[evtType]
	if !ok {
		return fmt.Errorf("%w %s", ErrUnsupportedContentType, evtType.Repr())
	}
	content.Parsed = reflect.New(structType).Interface()
	return json.Unmarshal(content.VeryRaw, &content.Parsed)
}

func mergeMaps(into, from map[string]any) {
	for key, newValue := range from {
		existingValue, ok := into[key]
		if !ok {
			into[key] = newValue
			continue
		}
		existingValueMap, okEx := existingValue.(map[string]any)
		newValueMap, okNew := newValue.(map[string]any)
		if okEx && okNew {
			mergeMaps(existingValueMap, newValueMap)
		} else {
			into[key] = newValue
		}
	}
}

func CastOrDefault[T any](content *Content) *T {
	casted, ok := content.Parsed.(*T)
	if ok {
		return casted
	}
	casted2, _ := content.Parsed.(T)
	return &casted2
}

// Helper cast functions below

func (content *Content) AsMember() *MemberEventContent {
	casted, ok := content.Parsed.(*MemberEventContent)
	if !ok {
		return &MemberEventContent{}
	}
	return casted
}
func (content *Content) AsPowerLevels() *PowerLevelsEventContent {
	casted, ok := content.Parsed.(*PowerLevelsEventContent)
	if !ok {
		return &PowerLevelsEventContent{}
	}
	return casted
}
func (content *Content) AsCanonicalAlias() *CanonicalAliasEventContent {
	casted, ok := content.Parsed.(*CanonicalAliasEventContent)
	if !ok {
		return &CanonicalAliasEventContent{}
	}
	return casted
}
func (content *Content) AsRoomName() *RoomNameEventContent {
	casted, ok := content.Parsed.(*RoomNameEventContent)
	if !ok {
		return &RoomNameEventContent{}
	}
	return casted
}
func (content *Content) AsRoomAvatar() *RoomAvatarEventContent {
	casted, ok := content.Parsed.(*RoomAvatarEventContent)
	if !ok {
		return &RoomAvatarEventContent{}
	}
	return casted
}
func (content *Content) AsTopic() *TopicEventContent {
	casted, ok := content.Parsed.(*TopicEventContent)
	if !ok {
		return &TopicEventContent{}
	}
	return casted
}
func (content *Content) AsTombstone() *TombstoneEventContent {
	casted, ok := content.Parsed.(*TombstoneEventContent)
	if !ok {
		return &TombstoneEventContent{}
	}
	return casted
}
func (content *Content) AsCreate() *CreateEventContent {
	casted, ok := content.Parsed.(*CreateEventContent)
	if !ok {
		return &CreateEventContent{}
	}
	return casted
}
func (content *Content) AsJoinRules() *JoinRulesEventContent {
	casted, ok := content.Parsed.(*JoinRulesEventContent)
	if !ok {
		return &JoinRulesEventContent{}
	}
	return casted
}
func (content *Content) AsHistoryVisibility() *HistoryVisibilityEventContent {
	casted, ok := content.Parsed.(*HistoryVisibilityEventContent)
	if !ok {
		return &HistoryVisibilityEventContent{}
	}
	return casted
}
func (content *Content) AsGuestAccess() *GuestAccessEventContent {
	casted, ok := content.Parsed.(*GuestAccessEventContent)
	if !ok {
		return &GuestAccessEventContent{}
	}
	return casted
}
func (content *Content) AsPinnedEvents() *PinnedEventsEventContent {
	casted, ok := content.Parsed.(*PinnedEventsEventContent)
	if !ok {
		return &PinnedEventsEventContent{}
	}
	return casted
}
func (content *Content) AsEncryption() *EncryptionEventContent {
	casted, ok := content.Parsed.(*EncryptionEventContent)
	if !ok {
		return &EncryptionEventContent{}
	}
	return casted
}
func (content *Content) AsBridge() *BridgeEventContent {
	casted, ok := content.Parsed.(*BridgeEventContent)
	if !ok {
		return &BridgeEventContent{}
	}
	return casted
}
func (content *Content) AsSpaceChild() *SpaceChildEventContent {
	casted, ok := content.Parsed.(*SpaceChildEventContent)
	if !ok {
		return &SpaceChildEventContent{}
	}
	return casted
}
func (content *Content) AsSpaceParent() *SpaceParentEventContent {
	casted, ok := content.Parsed.(*SpaceParentEventContent)
	if !ok {
		return &SpaceParentEventContent{}
	}
	return casted
}
func (content *Content) AsElementFunctionalMembers() *ElementFunctionalMembersContent {
	casted, ok := content.Parsed.(*ElementFunctionalMembersContent)
	if !ok {
		return &ElementFunctionalMembersContent{}
	}
	return casted
}
func (content *Content) AsMessage() *MessageEventContent {
	casted, ok := content.Parsed.(*MessageEventContent)
	if !ok {
		return &MessageEventContent{}
	}
	return casted
}
func (content *Content) AsEncrypted() *EncryptedEventContent {
	casted, ok := content.Parsed.(*EncryptedEventContent)
	if !ok {
		return &EncryptedEventContent{}
	}
	return casted
}
func (content *Content) AsRedaction() *RedactionEventContent {
	casted, ok := content.Parsed.(*RedactionEventContent)
	if !ok {
		return &RedactionEventContent{}
	}
	return casted
}
func (content *Content) AsReaction() *ReactionEventContent {
	casted, ok := content.Parsed.(*ReactionEventContent)
	if !ok {
		return &ReactionEventContent{}
	}
	return casted
}
func (content *Content) AsTag() *TagEventContent {
	casted, ok := content.Parsed.(*TagEventContent)
	if !ok {
		return &TagEventContent{}
	}
	return casted
}
func (content *Content) AsDirectChats() *DirectChatsEventContent {
	casted, ok := content.Parsed.(*DirectChatsEventContent)
	if !ok {
		return &DirectChatsEventContent{}
	}
	return casted
}
func (content *Content) AsFullyRead() *FullyReadEventContent {
	casted, ok := content.Parsed.(*FullyReadEventContent)
	if !ok {
		return &FullyReadEventContent{}
	}
	return casted
}
func (content *Content) AsIgnoredUserList() *IgnoredUserListEventContent {
	casted, ok := content.Parsed.(*IgnoredUserListEventContent)
	if !ok {
		return &IgnoredUserListEventContent{}
	}
	return casted
}
func (content *Content) AsMarkedUnread() *MarkedUnreadEventContent {
	casted, ok := content.Parsed.(*MarkedUnreadEventContent)
	if !ok {
		return &MarkedUnreadEventContent{}
	}
	return casted
}
func (content *Content) AsTyping() *TypingEventContent {
	casted, ok := content.Parsed.(*TypingEventContent)
	if !ok {
		return &TypingEventContent{}
	}
	return casted
}
func (content *Content) AsReceipt() *ReceiptEventContent {
	casted, ok := content.Parsed.(*ReceiptEventContent)
	if !ok {
		return &ReceiptEventContent{}
	}
	return casted
}
func (content *Content) AsPresence() *PresenceEventContent {
	casted, ok := content.Parsed.(*PresenceEventContent)
	if !ok {
		return &PresenceEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKey() *RoomKeyEventContent {
	casted, ok := content.Parsed.(*RoomKeyEventContent)
	if !ok {
		return &RoomKeyEventContent{}
	}
	return casted
}
func (content *Content) AsForwardedRoomKey() *ForwardedRoomKeyEventContent {
	casted, ok := content.Parsed.(*ForwardedRoomKeyEventContent)
	if !ok {
		return &ForwardedRoomKeyEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKeyRequest() *RoomKeyRequestEventContent {
	casted, ok := content.Parsed.(*RoomKeyRequestEventContent)
	if !ok {
		return &RoomKeyRequestEventContent{}
	}
	return casted
}
func (content *Content) AsRoomKeyWithheld() *RoomKeyWithheldEventContent {
	casted, ok := content.Parsed.(*RoomKeyWithheldEventContent)
	if !ok {
		return &RoomKeyWithheldEventContent{}
	}
	return casted
}
func (content *Content) AsBeeperStreamSubscribe() *BeeperStreamSubscribeEventContent {
	casted, ok := content.Parsed.(*BeeperStreamSubscribeEventContent)
	if !ok {
		return &BeeperStreamSubscribeEventContent{}
	}
	return casted
}

func (content *Content) AsBeeperStreamUpdate() *BeeperStreamUpdateEventContent {
	casted, ok := content.Parsed.(*BeeperStreamUpdateEventContent)
	if !ok {
		return &BeeperStreamUpdateEventContent{}
	}
	return casted
}

func (content *Content) AsCallInvite() *CallInviteEventContent {
	casted, ok := content.Parsed.(*CallInviteEventContent)
	if !ok {
		return &CallInviteEventContent{}
	}
	return casted
}
func (content *Content) AsCallCandidates() *CallCandidatesEventContent {
	casted, ok := content.Parsed.(*CallCandidatesEventContent)
	if !ok {
		return &CallCandidatesEventContent{}
	}
	return casted
}
func (content *Content) AsCallAnswer() *CallAnswerEventContent {
	casted, ok := content.Parsed.(*CallAnswerEventContent)
	if !ok {
		return &CallAnswerEventContent{}
	}
	return casted
}
func (content *Content) AsCallReject() *CallRejectEventContent {
	casted, ok := content.Parsed.(*CallRejectEventContent)
	if !ok {
		return &CallRejectEventContent{}
	}
	return casted
}
func (content *Content) AsCallSelectAnswer() *CallSelectAnswerEventContent {
	casted, ok := content.Parsed.(*CallSelectAnswerEventContent)
	if !ok {
		return &CallSelectAnswerEventContent{}
	}
	return casted
}
func (content *Content) AsCallNegotiate() *CallNegotiateEventContent {
	casted, ok := content.Parsed.(*CallNegotiateEventContent)
	if !ok {
		return &CallNegotiateEventContent{}
	}
	return casted
}
func (content *Content) AsCallHangup() *CallHangupEventContent {
	casted, ok := content.Parsed.(*CallHangupEventContent)
	if !ok {
		return &CallHangupEventContent{}
	}
	return casted
}
func (content *Content) AsModPolicy() *ModPolicyContent {
	casted, ok := content.Parsed.(*ModPolicyContent)
	if !ok {
		return &ModPolicyContent{}
	}
	return casted
}
func (content *Content) AsVerificationRequest() *VerificationRequestEventContent {
	casted, ok := content.Parsed.(*VerificationRequestEventContent)
	if !ok {
		return &VerificationRequestEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationReady() *VerificationReadyEventContent {
	casted, ok := content.Parsed.(*VerificationReadyEventContent)
	if !ok {
		return &VerificationReadyEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationStart() *VerificationStartEventContent {
	casted, ok := content.Parsed.(*VerificationStartEventContent)
	if !ok {
		return &VerificationStartEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationDone() *VerificationDoneEventContent {
	casted, ok := content.Parsed.(*VerificationDoneEventContent)
	if !ok {
		return &VerificationDoneEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationCancel() *VerificationCancelEventContent {
	casted, ok := content.Parsed.(*VerificationCancelEventContent)
	if !ok {
		return &VerificationCancelEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationAccept() *VerificationAcceptEventContent {
	casted, ok := content.Parsed.(*VerificationAcceptEventContent)
	if !ok {
		return &VerificationAcceptEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationKey() *VerificationKeyEventContent {
	casted, ok := content.Parsed.(*VerificationKeyEventContent)
	if !ok {
		return &VerificationKeyEventContent{}
	}
	return casted
}
func (content *Content) AsVerificationMAC() *VerificationMACEventContent {
	casted, ok := content.Parsed.(*VerificationMACEventContent)
	if !ok {
		return &VerificationMACEventContent{}
	}
	return casted
}
