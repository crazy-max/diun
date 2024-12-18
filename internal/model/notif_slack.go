package model

import "github.com/crazy-max/diun/v4/pkg/utl"

// NotifSlackDefaultTemplateBody ...
const NotifSlackDefaultTemplateBody = "<!channel> Docker tag {{ if .Entry.Image.HubLink }}<{{ .Entry.Image.HubLink }}|`{{ .Entry.Image }}`>{{ else }}`{{ .Entry.Image }}`{{ end }}  {{ if (eq .Entry.Status \"new\") }}available{{ else }}updated{{ end }}."

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	WebhookURL     string `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"omitempty"`
	WebhookURLFile string `yaml:"webhookURLFile,omitempty" json:"webhookURLFile,omitempty" validate:"omitempty,file"`
	RenderFields   *bool  `yaml:"renderFields,omitempty" json:"renderFields,omitempty" validate:"required"`
	TemplateBody   string `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifSlack) GetDefaults() *NotifSlack {
	n := &NotifSlack{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifSlack) SetDefaults() {
	s.RenderFields = utl.NewTrue()
	s.TemplateBody = NotifSlackDefaultTemplateBody
}
