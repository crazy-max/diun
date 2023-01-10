package model

import "github.com/crazy-max/diun/v4/pkg/utl"

type APIMetrics struct {
	EnableAPI  *bool  `yaml:"enableApi,omitempty" json:"enableApi,omitempty"`
	EnableScan *bool  `yaml:"enableScan,omitempty" json:"enableScan,omitempty"`
	Port       string `yaml:"port,omitempty" json:"port,omitempty"`
	Token      string `yaml:"token,omitempty" json:"token,omitempty"`
	APIPath    string `yaml:"apiPath,omitempty" json:"apiPath,omitempty"`
	ScanPath   string `yaml:"scanPath,omitempty" json:"scanPath,omitempty"`
}

// GetDefaults gets the default values
func (s *APIMetrics) GetDefaults() *APIMetrics {
	n := &APIMetrics{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *APIMetrics) SetDefaults() {
	s.EnableAPI = utl.NewFalse()
	s.EnableScan = utl.NewFalse()
	s.APIPath = "/v1/metrics"
	s.ScanPath = "/v1/scan"
	s.Port = "6080"
	s.Token = "ApiToken"
}
