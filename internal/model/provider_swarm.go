package model

import (
	"github.com/crazy-max/diun/v3/pkg/utl"
)

// PrdSwarm holds swarm provider configuration
type PrdSwarm struct {
	Endpoint       string `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"omitempty"`
	APIVersion     string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty" validate:"omitempty"`
	TLSCertsPath   string `yaml:"tlsCertsPath,omitempty" json:"tlsCertsPath,omitempty" validate:"omitempty"`
	TLSVerify      *bool  `yaml:"tlsVerify,omitempty" json:"tlsVerify,omitempty" validate:"required"`
	WatchByDefault *bool  `yaml:"watchByDefault,omitempty" json:"watchByDefault,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *PrdSwarm) GetDefaults() *PrdSwarm {
	n := &PrdSwarm{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *PrdSwarm) SetDefaults() {
	s.TLSVerify = utl.NewTrue()
	s.WatchByDefault = utl.NewFalse()
}
