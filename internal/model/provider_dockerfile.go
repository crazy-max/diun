package model

// PrdDockerfile holds dockerfile provider configuration
type PrdDockerfile struct {
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *PrdDockerfile) GetDefaults() *PrdDockerfile {
	return nil
}

// SetDefaults sets the default values
func (s *PrdDockerfile) SetDefaults() {
	// noop
}
