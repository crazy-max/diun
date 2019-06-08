package model

// Watch holds data necessary for watch configuration
type Watch struct {
	Schedule string `yaml:"schedule,omitempty"`
	Os       string `yaml:"os,omitempty"`
	Arch     string `yaml:"arch,omitempty"`
}
