package model

import "github.com/crazy-max/diun/v4/pkg/utl"

type APIMetrics struct {
	Port   string `yaml:"port,omitempty" json:"port,omitempty"`
	Token  string `yaml:"token,omitempty" json:"token,omitempty"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Enable *bool  `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// GetDefaults gets the default values
func (s *APIMetrics) GetDefaults() *APIMetrics {
	n := &APIMetrics{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *APIMetrics) SetDefaults() {
	s.Enable = utl.NewFalse()
	s.Path = "/v1/metrics"
	s.Port = "6080"
	s.Token = "ApiToken"
}
