package model

import "github.com/crazy-max/diun/v4/pkg/utl"

// NotifTeamsDefaultTemplateBody ...
const NotifTeamsDefaultTemplateBody = "Docker tag {{ if .Entry.Image.HubLink }}[`{{ .Entry.Image }}`]({{ .Entry.Image.HubLink }}){{ else }}`{{ .Entry.Image }}`{{ end }} {{ if (eq .Entry.Status \"new\") }}available{{ else }}updated{{ end }}."

// NotifTeams holds Teams notification configuration details
type NotifTeams struct {
	WebhookURL   string `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
	RenderFacts  *bool  `yaml:"renderFacts,omitempty" json:"renderFacts,omitempty" validate:"required"`
	TemplateBody string `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifTeams) GetDefaults() *NotifTeams {
	n := &NotifTeams{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifTeams) SetDefaults() {
	s.RenderFacts = utl.NewTrue()
	s.TemplateBody = NotifTeamsDefaultTemplateBody
}
