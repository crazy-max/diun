package model

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	WebhookURL string `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifSlack) GetDefaults() *NotifSlack {
	return nil
}

// SetDefaults sets the default values
func (s *NotifSlack) SetDefaults() {
	// noop
}
