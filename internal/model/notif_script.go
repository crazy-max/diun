package model

// NotifScript holds script notification configuration details
type NotifScript struct {
	Cmd  string   `yaml:"cmd,omitempty" json:"cmd,omitempty" validate:"required"`
	Args []string `yaml:"args,omitempty" json:"args,omitempty" validate:"omitempty"`
	Dir  string   `yaml:"dir,omitempty" json:"dir,omitempty" validate:"omitempty,dir"`
}

// GetDefaults gets the default values
func (s *NotifScript) GetDefaults() *NotifScript {
	return nil
}

// SetDefaults sets the default values
func (s *NotifScript) SetDefaults() {
	// noop
}
