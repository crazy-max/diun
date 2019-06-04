package model

// Watch holds data necessary for watch configuration
type Watch struct {
	Schedule string `yaml:"schedule,omitempty"`
}
