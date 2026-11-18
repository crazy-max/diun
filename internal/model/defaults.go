package model

import (
	"github.com/crazy-max/diun/v4/pkg/registry"
)

// Defaults holds data necessary for image defaults configuration
type Defaults struct {
	WatchMode           registry.WatchMode `yaml:"watchMode,omitempty" json:"watchMode,omitempty" validate:"omitempty,oneof=tag all semver semver+prereleases"`
	WatchRepo           *bool              `yaml:"watchRepo,omitempty" json:"watchRepo,omitempty"` // Deprecated: use WatchMode instead.
	NotifyOn            []NotifyOn         `yaml:"notifyOn,omitempty" json:"notifyOn,omitempty"`
	MaxTags             int                `yaml:"maxTags,omitempty" json:"maxTags,omitempty"`
	SortTags            registry.SortTag   `yaml:"sortTags,omitempty" json:"sortTags,omitempty"`
	IncludeTags         []string           `yaml:"includeTags,omitempty" json:"includeTags,omitempty"`
	ExcludeTags         []string           `yaml:"excludeTags,omitempty" json:"excludeTags,omitempty"`
	AssumeImmutableTags *bool              `yaml:"assumeImmutableTags,omitempty" json:"assumeImmutableTags,omitempty"`
	Metadata            map[string]string  `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// GetDefaults gets the default values
func (s *Defaults) GetDefaults() *Defaults {
	n := &Defaults{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *Defaults) SetDefaults() {
	s.NotifyOn = NotifyOnDefaults
	s.SortTags = registry.SortTagReverse
}
