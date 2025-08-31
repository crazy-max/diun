package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifPushover holds Pushover notification configuration details
type NotifPushover struct {
	Token         string         `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile     string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Recipient     string         `yaml:"recipient,omitempty" json:"recipient,omitempty" validate:"omitempty"`
	RecipientFile string         `yaml:"recipientFile,omitempty" json:"recipientFile,omitempty" validate:"omitempty,file"`
	Priority      int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=-2,max=2"`
	Sound         string         `yaml:"sound,omitempty" json:"sound,omitempty" validate:"omitempty"`
	Timeout       *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateTitle string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody  string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifPushover) GetDefaults() *NotifPushover {
	n := &NotifPushover{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifPushover) SetDefaults() {
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifDefaultTemplateBody
}
