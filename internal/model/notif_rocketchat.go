package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifRocketChatDefaultTemplateBody ...
const NotifRocketChatDefaultTemplateBody = `Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }} on {{ .Meta.Hostname }}.`

// NotifRocketChat holds Rocket.Chat notification configuration details
type NotifRocketChat struct {
	Endpoint      string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Channel       string         `yaml:"channel,omitempty" json:"channel,omitempty" validate:"required"`
	UserID        string         `yaml:"userID,omitempty" json:"userID,omitempty" validate:"required"`
	Token         string         `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile     string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Timeout       *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateTitle string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody  string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifRocketChat) GetDefaults() *NotifRocketChat {
	n := &NotifRocketChat{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifRocketChat) SetDefaults() {
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifRocketChatDefaultTemplateBody
}
