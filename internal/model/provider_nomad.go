package model

import (
	"github.com/crazy-max/diun/v4/pkg/utl"
)

// PrdNomad holds nomad provider configuration
type PrdNomad struct {
	Address        string `yaml:"address" json:"address,omitempty" validate:"omitempty"`
	Region         string `yaml:"region,omitempty" json:"region,omitempty" validate:"omitempty"`
	SecretID       string `yaml:"secretID,omitempty" json:"secretID,omitempty" validate:"omitempty"`
	Namespace      string `yaml:"namespace,omitempty" json:"namespace,omitempty" validate:"omitempty"`
	TLSInsecure    *bool  `yaml:"tlsInsecure" json:"tlsInsecure,omitempty" validate:"required"`
	WatchByDefault *bool  `yaml:"watchByDefault" json:"watchByDefault,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *PrdNomad) GetDefaults() *PrdNomad {
	n := &PrdNomad{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *PrdNomad) SetDefaults() {
	s.TLSInsecure = utl.NewFalse()
	s.WatchByDefault = utl.NewFalse()
}
