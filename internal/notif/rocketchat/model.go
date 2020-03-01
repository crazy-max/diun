package rocketchat

import "encoding/json"

// Message contains all the information for a message
type Message struct {
	Alias       string       `json:"alias,omitempty"`
	Avatar      string       `json:"avatar,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Emoji       string       `json:"emoji,omitempty"`
	RoomID      string       `json:"roomId,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment contains all the information for an attachment
type Attachment struct {
	AudioURL    string            `json:"audio_url,omitempty"`
	AuthorIcon  string            `json:"author_icon,omitempty"`
	AuthorLink  string            `json:"author_link,omitempty"`
	AuthorName  string            `json:"author_name,omitempty"`
	Collapsed   bool              `json:"collapsed,omitempty"`
	Color       bool              `json:"color,omitempty"`
	Fields      []AttachmentField `json:"fields,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	MessageLink string            `json:"message_link,omitempty"`
	Text        string            `json:"text"`
	ThumbURL    string            `json:"thumb_url,omitempty"`
	Title       string            `json:"title,omitempty"`
	TitleLink   string            `json:"title_link,omitempty"`
	Ts          json.Number       `json:"ts,omitempty"`
	VideoURL    string            `json:"video_url,omitempty"`
}

// AttachmentField contains information for an attachment field
// An Attachment can contain multiple of these
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
