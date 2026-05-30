package model

// PrdContainerd holds containerd provider configuration
type PrdContainerd struct {
	Endpoint       string   `yaml:"endpoint" json:"endpoint,omitempty" validate:"omitempty"`
	Namespaces     []string `yaml:"namespaces" json:"namespaces,omitempty" validate:"omitempty,dive,required"`
	WatchByDefault *bool    `yaml:"watchByDefault" json:"watchByDefault,omitempty" validate:"required"`
	WatchStopped   *bool    `yaml:"watchStopped" json:"watchStopped,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *PrdContainerd) GetDefaults() *PrdContainerd {
	n := &PrdContainerd{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *PrdContainerd) SetDefaults() {
	s.Namespaces = []string{"default"}
	s.WatchByDefault = new(false)
	s.WatchStopped = new(false)
}
