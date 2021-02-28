package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifGotify holds gotify notification configuration details
type NotifGotify struct {
	Endpoint  string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Token     string         `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Priority  int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=0"`
	Timeout   *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifGotify) GetDefaults() *NotifGotify {
	n := &NotifGotify{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifGotify) SetDefaults() {
	s.Priority = 1
	s.Timeout = utl.NewDuration(10 * time.Second)
}
