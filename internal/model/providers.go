package model

// Providers represents a provider configuration
type Providers struct {
	Docker     *PrdDocker     `yaml:"docker,omitempty" json:"docker,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Swarm      *PrdSwarm      `yaml:"swarm,omitempty" json:"swarm,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Kubernetes *PrdKubernetes `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty" label:"allowEmpty" file:"allowEmpty"`
	File       *PrdFile       `yaml:"file,omitempty" json:"file,omitempty"`
}

// GetDefaults gets the default values
func (s *Providers) GetDefaults() *Providers {
	return nil
}

// SetDefaults sets the default values
func (s *Providers) SetDefaults() {
	// noop
}
