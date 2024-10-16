package model

type NotifHomeAssistant struct {
    Scheme            string `yaml:"scheme,omitempty" json:"scheme,omitempty" validate:"required,oneof=mqtt mqtts ws wss"`
    Host              string `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
    Port              int    `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
    Username          string `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
    UsernameFile      string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
    Password          string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
    PasswordFile      string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
    Client            string `yaml:"client,omitempty" json:"client,omitempty" validate:"required"`
    DiscoveryPrefix   string `yaml:"discoveryPrefix,omitempty" json:"discoveryPrefix,omitempty" validate:"required"`
    Component         string `yaml:"component,omitempty" json:"component,omitempty" validate:"required"`
    NodeName       string `yaml:"nodeName,omitempty" json:"nodeName,omitempty" validate:"omitempty"`
    QoS               int    `yaml:"qos,omitempty" json:"qos,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifHomeAssistant) GetDefaults() *NotifHomeAssistant {
    n := &NotifHomeAssistant{}
    n.SetDefaults()
    return n
}

// SetDefaults sets the default values
func (s *NotifHomeAssistant) SetDefaults() {
    s.Scheme = "mqtt"
    s.Host = "localhost"
    s.Port = 1883
    s.DiscoveryPrefix = "homeassistant"
    s.Component = "sensor"
    s.NodeName = "diun"
    s.QoS = 0
}