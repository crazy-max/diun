package model

// NotifTelegram holds Telegram notification configuration details
type NotifTelegram struct {
	Token       string  `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile   string  `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	ChatIDs     []int64 `yaml:"chatIDs,omitempty" json:"chatIDs,omitempty" validate:"omitempty"`
	ChatIDsFile string  `yaml:"chatIDsFile,omitempty" json:"chatIDsFile,omitempty" validate:"omitempty,file"`
}

// GetDefaults gets the default values
func (s *NotifTelegram) GetDefaults() *NotifTelegram {
	return nil
}

// SetDefaults sets the default values
func (s *NotifTelegram) SetDefaults() {
	// noop
}
