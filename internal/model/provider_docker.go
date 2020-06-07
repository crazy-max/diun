package model

import (
	"github.com/crazy-max/diun/v3/pkg/utl"
)

// PrdDocker holds docker provider configuration
type PrdDocker struct {
	Endpoint       string `yaml:"endpoint" json:"endpoint,omitempty" validate:"omitempty"`
	APIVersion     string `yaml:"apiVersion" json:"apiVersion,omitempty" validate:"omitempty"`
	TLSCertsPath   string `yaml:"tlsCertsPath" json:"tlsCertsPath,omitempty" validate:"omitempty"`
	TLSVerify      *bool  `yaml:"tlsVerify" json:"tlsVerify,omitempty" validate:"required"`
	WatchByDefault *bool  `yaml:"watchByDefault" json:"watchByDefault,omitempty" validate:"required"`
	WatchStopped   *bool  `yaml:"watchStopped" json:"watchStopped,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *PrdDocker) GetDefaults() *PrdDocker {
	n := &PrdDocker{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *PrdDocker) SetDefaults() {
	s.TLSVerify = utl.NewTrue()
	s.WatchByDefault = utl.NewFalse()
	s.WatchStopped = utl.NewFalse()
}
