package model

// NotifPushover holds Pushover notification configuration details
type NotifPushover struct {
	Token     string `yaml:"token,omitempty" json:"token,omitempty" validate:"required"`
	Recipient string `yaml:"recipient,omitempty" json:"recipient,omitempty" validate:"required"`
	Priority  int    `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=-2,max=2"`
}

// GetDefaults gets the default values
func (s *NotifPushover) GetDefaults() *NotifPushover {
	return nil
}

// SetDefaults sets the default values
func (s *NotifPushover) SetDefaults() {
	// noop
}
