package model

import "github.com/crazy-max/diun/internal/model/provider"

// Providers represents a provider configuration
type Providers struct {
	Image []provider.Image `yaml:"image,omitempty" json:",omitempty"`
}
