package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifApprise holds apprise notification configuration details
type NotifApprise struct {
	Endpoint      string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Tags          []string       `yaml:"tags,omitempty" json:"tags,omitempty" validate:"required"`
	Timeout       *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateTitle string         `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody  string         `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifApprise) GetDefaults() *NotifApprise {
	n := &NotifApprise{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifApprise) SetDefaults() {
	s.Endpoint = "http://apprise:8000/notify/apprise"
	s.Tags = []string{"all"}
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifDefaultTemplateBody
}
