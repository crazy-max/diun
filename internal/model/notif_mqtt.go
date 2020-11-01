package model

type NotifMqtt struct {
	Host         string `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port         int    `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
	Username     string `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password     string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	Client       string `yaml:"client,omitempty" json:"client,omitempty" validate:"required"`
	Topic        string `yaml:"topic,omitempty" json:"topic,omitempty" validate:"required"`
	QoS          int    `yaml:"qos,omitempty" json:"qos,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifMqtt) GetDefaults() *NotifMqtt {
	n := &NotifMqtt{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifMqtt) SetDefaults() {
	s.Host = "localhost"
	s.Port = 1883
	s.QoS = 0
}
