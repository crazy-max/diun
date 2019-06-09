package model

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers  int    `yaml:"int,omitempty"`
	Schedule string `yaml:"schedule,omitempty"`
	Os       string `yaml:"os,omitempty"`
	Arch     string `yaml:"arch,omitempty"`
}
