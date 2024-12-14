package gotgbot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetLink is a helper method to easily get the message link (It will return an empty string in case of private or group chat type).
func (m Message) GetLink() string {
	if m.Chat.Type == "private" || m.Chat.Type == "group" {
		return ""
	}
	if m.Chat.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", m.Chat.Username, m.MessageId)
	}
	// Message links use raw chatIds without the -100 prefix; this trims that prefix.
	rawChatId := strings.TrimPrefix(strconv.FormatInt(m.Chat.Id, 10), "-100")
	return fmt.Sprintf("https://t.me/c/%s/%d", rawChatId, m.MessageId)
}

// GetText returns the message text, for both text messages and media messages. (Why is this not the telegram default!)
func (m Message) GetText() string {
	if m.Caption != "" {
		return m.Caption
	}
	return m.Text
}

// GetEntities returns the message entities, for both text messages and media messages. (Why is this not the telegram default!)
func (m Message) GetEntities() []MessageEntity {
	if len(m.CaptionEntities) > 0 {
		return m.CaptionEntities
	}
	return m.Entities
}

// Reply is a helper function to easily call Bot.SendMessage as a reply to an existing Message.
func (m Message) Reply(b *Bot, text string, opts *SendMessageOpts) (*Message, error) {
	if opts == nil {
		opts = &SendMessageOpts{}
	}

	if opts.ReplyParameters == nil || opts.ReplyParameters.MessageId == 0 {
		if opts.ReplyParameters == nil {
			opts.ReplyParameters = &ReplyParameters{}
		}
		opts.ReplyParameters.MessageId = m.MessageId
	}

	return b.SendMessage(m.Chat.Id, text, opts)
}

// Reply is a helper function to easily call Bot.SendMessage as a reply to an existing InaccessibleMessage.
func (im InaccessibleMessage) Reply(b *Bot, text string, opts *SendMessageOpts) (*Message, error) {
	if opts == nil {
		opts = &SendMessageOpts{}
	}

	if opts.ReplyParameters == nil || opts.ReplyParameters.MessageId == 0 {
		if opts.ReplyParameters == nil {
			opts.ReplyParameters = &ReplyParameters{}
		}
		opts.ReplyParameters.MessageId = im.MessageId
	}

	return b.SendMessage(im.Chat.Id, text, opts)
}

// ToMessage is a helper function to simplify dealing with telegram's message nonsense.
// It populates a standard message object with all of InaccessibleMessage's shared fields.
func (im InaccessibleMessage) ToMessage() *Message {
	return &Message{
		MessageId: im.MessageId,
		Date:      im.Date,
		Chat:      im.Chat,
	}
}

// ToChat is a helper function to turn a ChatFullInfo struct into a Chat.
func (c ChatFullInfo) ToChat() Chat {
	return Chat{
		Id:        c.Id,
		Type:      c.Type,
		Title:     c.Title,
		Username:  c.Username,
		FirstName: c.FirstName,
		LastName:  c.LastName,
		IsForum:   c.IsForum,
	}
}

// SendMessage is a helper function to easily call Bot.SendMessage in a chat.
func (c Chat) SendMessage(b *Bot, text string, opts *SendMessageOpts) (*Message, error) {
	return b.SendMessage(c.Id, text, opts)
}

// Unban is a helper function to easily call Bot.UnbanChatMember in a chat.
func (c Chat) Unban(b *Bot, userId int64, opts *UnbanChatMemberOpts) (bool, error) {
	return b.UnbanChatMember(c.Id, userId, opts)
}

// Promote is a helper function to easily call Bot.PromoteChatMember in a chat.
func (c Chat) Promote(b *Bot, userId int64, opts *PromoteChatMemberOpts) (bool, error) {
	return b.PromoteChatMember(c.Id, userId, opts)
}

// URL gets the URL the file can be downloaded from.
func (f File) URL(b *Bot, opts *RequestOpts) string {
	return b.FileURL(b.Token, f.FilePath, opts)
}

// IsJoinRequest returns true if ChatMemberUpdated originated from a join request; either from a direct join, or from an invitelink.
func (cm ChatMemberUpdated) IsJoinRequest() bool {
	return cm.ViaJoinRequest || (cm.InviteLink != nil && cm.InviteLink.CreatesJoinRequest)
}

// unmarshalMaybeInaccessibleMessage is a JSON unmarshal helper to marshal the right structs into a
// MaybeInaccessibleMessage interface based on the Date field.
// This method is manually maintained due to special-case handling on the Date field rather than a specific type field.
func unmarshalMaybeInaccessibleMessage(d json.RawMessage) (MaybeInaccessibleMessage, error) {
	if len(d) == 0 {
		return nil, nil
	}

	t := struct {
		Date int64
	}{}
	err := json.Unmarshal(d, &t)
	if err != nil {
		return nil, err
	}

	// As per the docs, date is always 0 for inaccessible messages:
	// https://core.telegram.org/bots/api#inaccessiblemessage
	if t.Date == 0 {
		s := InaccessibleMessage{}
		err := json.Unmarshal(d, &s)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	s := Message{}
	err = json.Unmarshal(d, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
