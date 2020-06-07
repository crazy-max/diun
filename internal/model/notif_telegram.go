package model

// NotifTelegram holds Telegram notification configuration details
type NotifTelegram struct {
	Token   string  `yaml:"token,omitempty" json:"token,omitempty" validate:"required"`
	ChatIDs []int64 `yaml:"chatIDs,omitempty" json:"chatIDs,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifTelegram) GetDefaults() *NotifTelegram {
	return nil
}

// SetDefaults sets the default values
func (s *NotifTelegram) SetDefaults() {
	// noop
}
