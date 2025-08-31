package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifRocketChatDefaultTemplateBody ...
const NotifRocketChatDefaultTemplateBody = `Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider {{ if (eq .Entry.Status "new") }}is available{{ else }}has been updated{{ end }} on {{ .Entry.Image.Domain }} registry (triggered by {{ .Meta.Hostname }} host).`

// NotifRocketChat holds Rocket.Chat notification configuration details
type NotifRocketChat struct {
	Endpoint         string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Channel          string         `yaml:"channel,omitempty" json:"channel,omitempty" validate:"required"`
	UserID           string         `yaml:"userID,omitempty" json:"userID,omitempty" validate:"required"`
	Token            string         `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile        string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	RenderAttachment *bool          `yaml:"renderAttachment,omitempty" json:"renderAttachment,omitempty" validate:"required"`
	Timeout          *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TLSSkipVerify    bool           `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" validate:"omitempty"`
	TLSCACertFiles   []string       `yaml:"tlsCaCertFiles,omitempty" json:"tlsCaCertFiles,omitempty" validate:"omitempty"`
	TemplateTitle    string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody     string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifRocketChat) GetDefaults() *NotifRocketChat {
	n := &NotifRocketChat{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifRocketChat) SetDefaults() {
	s.RenderAttachment = utl.NewTrue()
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifRocketChatDefaultTemplateBody
}
