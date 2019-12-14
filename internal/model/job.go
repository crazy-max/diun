package model

import (
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
)

// Job holds job configuration
type Job struct {
	Provider string
	Image    Image
	RegImage registry.Image
	Registry *docker.RegistryClient
}
