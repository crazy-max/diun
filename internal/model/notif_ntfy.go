package model

import (
	"time"
)

// NotifNtfy holds ntfy notification configuration details
type NotifNtfy struct {
	Endpoint       string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Token          string         `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile      string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Topic          string         `yaml:"topic,omitempty" json:"topic,omitempty" validate:"required"`
	Priority       int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=0"`
	Tags           []string       `yaml:"tags,omitempty" json:"tags,omitempty" validate:"required"`
	Icon           string         `yaml:"icon,omitempty" json:"icon,omitempty" validate:"omitempty"`
	Timeout        *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TLSSkipVerify  bool           `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" validate:"omitempty"`
	TLSCACertFiles []string       `yaml:"tlsCaCertFiles,omitempty" json:"tlsCaCertFiles,omitempty" validate:"omitempty"`
	TemplateTitle  string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody   string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifNtfy) GetDefaults() *NotifNtfy {
	n := &NotifNtfy{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifNtfy) SetDefaults() {
	s.Endpoint = "https://ntfy.sh"
	s.Priority = 3
	s.Tags = []string{"package"}
	s.Timeout = new(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifDefaultTemplateBody
}
