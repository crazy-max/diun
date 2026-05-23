package model

import (
	"time"
)

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers         int            `yaml:"workers,omitempty" json:"workers,omitempty" validate:"required,min=1"`
	Schedule        string         `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Jitter          *time.Duration `yaml:"jitter,omitempty" json:"jitter,omitempty" validate:"required"`
	FirstCheckNotif *bool          `yaml:"firstCheckNotif,omitempty" json:"firstCheckNotif,omitempty" validate:"required"`
	RunOnStartup    *bool          `yaml:"runOnStartup,omitempty" json:"runOnStartup,omitempty" validate:"required"`
	CompareDigest   *bool          `yaml:"compareDigest,omitempty" json:"compareDigest,omitempty" validate:"required"`
	Healthchecks    *Healthchecks  `yaml:"healthchecks,omitempty" json:"healthchecks,omitempty"`
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
	s.Jitter = new(30 * time.Second)
	s.FirstCheckNotif = new(false)
	s.RunOnStartup = new(true)
	s.CompareDigest = new(true)
}
