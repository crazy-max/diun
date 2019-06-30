package model

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers  int    `yaml:"workers,omitempty"`
	Schedule string `yaml:"schedule,omitempty"`
}
