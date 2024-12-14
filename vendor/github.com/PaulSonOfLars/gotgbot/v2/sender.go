package gotgbot

// Sender is a merge of the User and SenderChat fields of a message, to provide easier interaction with
// message senders from the telegram API.
type Sender struct {
	// The User defined as the sender (if applicable)
	User *User
	// The Chat defined as the sender (if applicable)
	Chat *Chat
	// Whether the sender was an automatic forward; eg, a linked channel.
	IsAutomaticForward bool
	// The location that was sent to. Required to determine if the sender is a linked channel, an anonymous channel,
	// or an anonymous admin.
	ChatId int64
	// The custom admin title of the anonymous group administrator sender.
	// Only available if IsAnonymousAdmin is true.
	AuthorSignature string
}

// GetSender populates the relevant fields of a Sender struct given a message.
func (m Message) GetSender() *Sender {
	return &Sender{
		User:               m.From,
		Chat:               m.SenderChat,
		IsAutomaticForward: m.IsAutomaticForward,
		ChatId:             m.Chat.Id,
		AuthorSignature:    m.AuthorSignature,
	}
}

// GetSender populates the relevant fields of a Sender struct given a reaction.
func (mru MessageReactionUpdated) GetSender() *Sender {
	return &Sender{
		User:   mru.User,
		Chat:   mru.ActorChat,
		ChatId: mru.Chat.Id,
	}
}

// GetSender populates the relevant fields of a Sender struct given a poll answer.
func (pa PollAnswer) GetSender() *Sender {
	return &Sender{
		User: pa.User,
		Chat: pa.VoterChat,
	}
}

// Id determines the sender ID.
// When a message is being sent by a chat/channel, telegram usually populates the User field with dummy values.
// For this reason, we prefer to return the Chat.Id if it is available, rather than a dummy User.Id.
func (s Sender) Id() int64 {
	if s.Chat != nil {
		return s.Chat.Id
	}
	if s.User != nil {
		return s.User.Id
	}
	return 0
}

// Username determines the sender username.
func (s Sender) Username() string {
	if s.Chat != nil {
		return s.Chat.Username
	}
	if s.User != nil {
		return s.User.Username
	}
	return ""
}

// Name determines the name of the sender.
// This is:
//   - Chat.Title for a Chat.
//   - User.FirstName + User.LastName for a User (the full name).
func (s Sender) Name() string {
	if s.Chat != nil {
		return s.Chat.Title
	}
	if s.User != nil {
		if s.User.LastName == "" {
			return s.User.FirstName
		}
		return s.User.FirstName + " " + s.User.LastName
	}
	return ""
}

// FirstName determines the firstname of the sender.
// This is:
//   - Chat.Title for a Chat.
//   - User.FirstName for a User.
func (s Sender) FirstName() string {
	if s.Chat != nil {
		return s.Chat.Title
	}
	if s.User != nil {
		return s.User.FirstName
	}
	return ""
}

// LastName determines the firstname of the sender.
// This is:
//   - empty for a Chat.
//   - User.LastName for a User.
func (s Sender) LastName() string {
	if s.Chat != nil {
		return "" // empty; we define the "title" as being a firstname, so there is no lastname.
	}
	if s.User != nil {
		return s.User.LastName
	}
	return ""
}

// IsUser returns true if the Sender is a User (including bot).
func (s Sender) IsUser() bool {
	return s.Chat == nil && s.User != nil
}

// IsBot returns true if the Sender is a bot.
// Returns false if the user is a bot setup by telegram for backwards compatibility with
// the sender_chat fields.
func (s Sender) IsBot() bool {
	return s.Chat == nil && s.User != nil && s.User.IsBot
}

// IsAnonymousAdmin returns true if the Sender is an anonymous admin sending to a group.
// For channel posts in a channel, see IsChannelPost.
func (s Sender) IsAnonymousAdmin() bool {
	return s.Chat != nil && s.Chat.Id == s.ChatId && s.Chat.Type != "channel"
}

// IsChannelPost returns true if the Sender is a channel admin posting to that same channel.
func (s Sender) IsChannelPost() bool {
	return s.Chat != nil && s.Chat.Id == s.ChatId && s.Chat.Type == "channel"
}

// IsAnonymousChannel returns true if the Sender is an anonymous channel sending to a group.
// For channel admins posting in their own channel, see IsChannelPost.
func (s Sender) IsAnonymousChannel() bool {
	return s.Chat != nil && s.Chat.Id != s.ChatId && !s.IsAutomaticForward && s.Chat.Type == "channel"
}

// IsLinkedChannel returns true if the Sender is a linked channel sending to the group it is linked to.
func (s Sender) IsLinkedChannel() bool {
	return s.Chat != nil && s.Chat.Id != s.ChatId && s.IsAutomaticForward
}
