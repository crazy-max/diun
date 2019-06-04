package model

// Webhook holds webhook notification configuration details
type Webhook struct {
	Enable   bool              `yaml:"enable,omitempty"`
	Endpoint string            `yaml:"endpoint,omitempty"`
	Method   string            `yaml:"method,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Timeout  int               `yaml:"timeout,omitempty"`
}
