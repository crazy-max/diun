package model

// Metrics holds data necessary for Prometheus metrics configuration.
type Metrics struct {
	Enabled   *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty" validate:"required"`
	Addr      string `yaml:"addr,omitempty" json:"addr,omitempty" validate:"required"`
	Path      string `yaml:"path,omitempty" json:"path,omitempty" validate:"required,startswith=/"`
	Token     string `yaml:"token,omitempty" json:"token,omitempty"`
	TokenFile string `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
}

// GetDefaults gets the default values.
func (s *Metrics) GetDefaults() *Metrics {
	n := &Metrics{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values.
func (s *Metrics) SetDefaults() {
	s.Enabled = new(false)
	s.Addr = ":9090"
	s.Path = "/metrics"
}
