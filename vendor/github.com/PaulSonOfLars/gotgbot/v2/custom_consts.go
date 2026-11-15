package gotgbot

// The consts listed below represent all the parse_mode options that can be sent to telegram.
const (
	ParseModeHTML       = "HTML"
	ParseModeMarkdownV2 = "MarkdownV2"
	ParseModeMarkdown   = "Markdown"
	ParseModeNone       = ""
)

// ChatMemberStatus contain some inconsistent naming schemes.
// We add aliases here for API clarity + backwards compatibility.
const (
	ChatMemberStatusOwner  = ChatMemberStatusCreator
	ChatMemberStatusBanned = ChatMemberStatusKicked
)
