package model

// NotifTeams holds Teams notification configuration details
type NotifTeams struct {
	WebhookURL string `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifTeams) GetDefaults() *NotifTeams {
	return nil
}

// SetDefaults sets the default values
func (s *NotifTeams) SetDefaults() {
	// noop
}
