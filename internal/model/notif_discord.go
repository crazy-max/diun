package model

import (
	"time"
)

// NotifDiscord holds Discord notification configuration details
type NotifDiscord struct {
	WebhookURL     string         `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"omitempty"`
	WebhookURLFile string         `yaml:"webhookURLFile,omitempty" json:"webhookURLFile,omitempty" validate:"omitempty,file"`
	Mentions       []string       `yaml:"mentions,omitempty" json:"mentions,omitempty"`
	RenderEmbeds   *bool          `yaml:"renderEmbeds,omitempty" json:"renderEmbeds,omitempty" validate:"required"`
	RenderFields   *bool          `yaml:"renderFields,omitempty" json:"renderFields,omitempty" validate:"required"`
	Timeout        *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	Proxy          string         `yaml:"proxy,omitempty" json:"proxy,omitempty" validate:"omitempty,url"`
	TemplateBody   string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifDiscord) GetDefaults() *NotifDiscord {
	n := &NotifDiscord{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifDiscord) SetDefaults() {
	s.RenderEmbeds = new(true)
	s.RenderFields = new(true)
	s.Timeout = new(10 * time.Second)
	s.TemplateBody = NotifDefaultTemplateBody
}
