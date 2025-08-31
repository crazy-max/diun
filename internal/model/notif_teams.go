package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifTeamsDefaultTemplateBody ...
const NotifTeamsDefaultTemplateBody = "Docker tag {{ if .Entry.Image.HubLink }}[`{{ .Entry.Image }}`]({{ .Entry.Image.HubLink }}){{ else }}`{{ .Entry.Image }}`{{ end }} {{ if (eq .Entry.Status \"new\") }}available{{ else }}updated{{ end }}."

// NotifTeams holds Teams notification configuration details
type NotifTeams struct {
	WebhookURL     string         `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"omitempty"`
	WebhookURLFile string         `yaml:"webhookURLFile,omitempty" json:"webhookURLFile,omitempty" validate:"omitempty,file"`
	RenderFacts    *bool          `yaml:"renderFacts,omitempty" json:"renderFacts,omitempty" validate:"required"`
	Timeout        *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TLSSkipVerify  bool           `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" validate:"omitempty"`
	TLSCACertFiles []string       `yaml:"tlsCaCertFiles,omitempty" json:"tlsCaCertFiles,omitempty" validate:"omitempty"`
	TemplateBody   string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifTeams) GetDefaults() *NotifTeams {
	n := &NotifTeams{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifTeams) SetDefaults() {
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.RenderFacts = utl.NewTrue()
	s.TemplateBody = NotifTeamsDefaultTemplateBody
}
