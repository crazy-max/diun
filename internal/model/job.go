package model

import "github.com/crazy-max/diun/pkg/docker"

// Job holds job configuration
type Job struct {
	Provider string
	ID       string
	Image    Image
	Registry *docker.RegistryClient
}
