package model

// NotifMatrix holds Matrix notification configuration details
type NotifMatrix struct {
	HomeserverURL string `yaml:"homeserverURL,omitempty" json:"homeserverURL,omitempty" validate:"required"`
	User          string `yaml:"user,omitempty" json:"user,omitempty" validate:"omitempty"`
	UserFile      string `yaml:"userFile,omitempty" json:"userFile,omitempty" validate:"omitempty,file"`
	Password      string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile  string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	RoomID        string `yaml:"roomID,omitempty" json:"roomID,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifMatrix) GetDefaults() *NotifMatrix {
	n := &NotifMatrix{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifMatrix) SetDefaults() {
	s.HomeserverURL = "https://matrix.org"
}
