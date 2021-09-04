package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifDiscord holds Discord notification configuration details
type NotifDiscord struct {
	WebhookURL   string         `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
	Mentions     []string       `yaml:"mentions,omitempty" json:"mentions,omitempty"`
	RenderFields *bool          `yaml:"renderFields,omitempty" json:"renderFields,omitempty" validate:"required"`
	Timeout      *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateBody string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifDiscord) GetDefaults() *NotifDiscord {
	n := &NotifDiscord{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifDiscord) SetDefaults() {
	s.RenderFields = utl.NewTrue()
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateBody = NotifDefaultTemplateBody
}
