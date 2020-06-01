package model

// NotifAmqp holds amqp notification configuration details
type NotifAmqp struct {
	Username     string `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password     string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	Host         string `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port         int    `yaml:"port,omitempty" json:"port,omitempty" validate:"required"`
	Queue        string `yaml:"queue,omitempty" json:"queue,omitempty" validate:"required"`
	Exchange     string `yaml:"exchange,omitempty" json:"exchange,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifAmqp) GetDefaults() *NotifAmqp {
	n := &NotifAmqp{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifAmqp) SetDefaults() {
	s.Host = "localhost"
	s.Port = 5672
}
