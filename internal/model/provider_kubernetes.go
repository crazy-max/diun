package model

import (
"github.com/crazy-max/diun/v4/pkg/utl"
)

// PrdKubernetes holds kubernetes provider configuration
type PrdKubernetes struct {
	Endpoint          string   `yaml:"endpoint" json:"endpoint,omitempty" validate:"omitempty"`
	Token             string   `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile         string   `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	CertAuthFilePath  string   `yaml:"certAuthFilePath" json:"certAuthFilePath,omitempty" validate:"omitempty"`
	TLSInsecure       *bool    `yaml:"tlsInsecure" json:"tlsInsecure,omitempty" validate:"required"`
	Namespaces        []string `yaml:"namespaces" json:"namespaces,omitempty" validate:"omitempty"`
	NamespacesExclude []string `yaml:"namespacesExclude" json:"namespacesExclude,omitempty" validate:"omitempty"`
	WatchByDefault    *bool    `yaml:"watchByDefault" json:"watchByDefault,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *PrdKubernetes) GetDefaults() *PrdKubernetes {
	n := &PrdKubernetes{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *PrdKubernetes) SetDefaults() {
	s.TLSInsecure = utl.NewFalse()
	s.WatchByDefault = utl.NewFalse()
}
