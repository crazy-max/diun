package model

// Healthchecks holds data necessary for Healthchecks configuration
type Healthchecks struct {
	BaseURL string `yaml:"baseURL,omitempty" json:"baseURL,omitempty"`
	UUID    string `yaml:"uuid,omitempty" json:"uuid,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *Healthchecks) GetDefaults() *Healthchecks {
	n := &Healthchecks{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *Healthchecks) SetDefaults() {
	// noop
}
