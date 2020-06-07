package model

import (
	"github.com/crazy-max/diun/v4/pkg/utl"
)

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers         int    `yaml:"workers,omitempty" json:"workers,omitempty" validate:"required,min=1"`
	Schedule        string `yaml:"schedule,omitempty" json:"schedule,omitempty" validate:"required"`
	FirstCheckNotif *bool  `yaml:"firstCheckNotif,omitempty" json:"firstCheckNotif,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *Watch) GetDefaults() *Watch {
	n := &Watch{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *Watch) SetDefaults() {
	s.Workers = 10
	s.Schedule = "0 * * * *"
	s.FirstCheckNotif = utl.NewFalse()
}
