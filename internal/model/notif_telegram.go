package model

// NotifTelegramDefaultTemplateBody ...
const NotifTelegramDefaultTemplateBody = `Docker tag {{ if .Entry.Image.HubLink }}[{{ .Entry.Image }}]({{ .Entry.Image.HubLink }}){{ else }}{{ .Entry.Image }}{{ end }} which you subscribed to through {{ .Entry.Provider }} provider {{ if (eq .Entry.Status "new") }}is available{{ else }}has been updated{{ end }} on {{ .Entry.Image.Domain }} registry (triggered by {{ escapeMarkdown .Meta.Hostname }} host).`

// NotifTelegram holds Telegram notification configuration details
type NotifTelegram struct {
	Token               string   `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile           string   `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	ChatIDs             []string `yaml:"chatIDs,omitempty" json:"chatIDs,omitempty" validate:"omitempty"`
	ChatIDsFile         string   `yaml:"chatIDsFile,omitempty" json:"chatIDsFile,omitempty" validate:"omitempty,file"`
	TemplateBody        string   `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
	DisableNotification *bool    `yaml:"disableNotification,omitempty" json:"disableNotification,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifTelegram) GetDefaults() *NotifTelegram {
	n := &NotifTelegram{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifTelegram) SetDefaults() {
	s.TemplateBody = NotifTelegramDefaultTemplateBody
}
