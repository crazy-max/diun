package model

import (
	"github.com/crazy-max/diun/pkg/registry"
)

// Job holds job configuration
type Job struct {
	Provider   string
	Image      Image
	RegImage   registry.Image
	Registry   *registry.Client
	FirstCheck bool
}
