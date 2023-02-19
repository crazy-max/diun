package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifNtfy holds ntfy notification configuration details
type NotifNtfy struct {
	Endpoint      string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Topic         string         `yaml:"topic,omitempty" json:"topic,omitempty" validate:"required"`
	Priority      int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=0"`
	Tags          []string       `yaml:"tags,omitempty" json:"tags,omitempty" validate:"required"`
	Timeout       *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateTitle string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody  string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
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
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifDefaultTemplateBody
}
