package model

import (
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/utl"
)

// ImageDefaults holds data necessary for image defaults configuration
type ImageDefaults struct {
	WatchRepo   *bool             `yaml:"watchRepo,omitempty" json:"watchRepo,omitempty"`
	NotifyOn    []NotifyOn        `yaml:"notifyOn,omitempty" json:"notifyOn,omitempty"`
	MaxTags     int               `yaml:"maxTags,omitempty" json:"maxTags,omitempty"`
	SortTags    registry.SortTag  `yaml:"sortTags,omitempty" json:"sortTags,omitempty"`
	IncludeTags []string          `yaml:"includeTags,omitempty" json:"includeTags,omitempty"`
	ExcludeTags []string          `yaml:"excludeTags,omitempty" json:"excludeTags,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// GetDefaults gets the default values
func (s *ImageDefaults) GetDefaults() *Watch {
	n := &Watch{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *ImageDefaults) SetDefaults() {
	s.WatchRepo = utl.NewFalse()
	s.NotifyOn = NotifyOnDefaults
	s.SortTags = registry.SortTagReverse
}
