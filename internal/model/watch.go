package model

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers             int    `yaml:"workers,omitempty"`
	Schedule            string `yaml:"schedule,omitempty"`
	Docker              bool   `yaml:"docker,omitempty"`
	UnlabeledContainers bool   `yaml:"unlabeled-containers,omitempty"`
	StoppedContainers   bool   `yaml:"stopped-containers,omitempty"`
}
