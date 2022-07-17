package model

import "github.com/crazy-max/diun/v4/pkg/registry"

// Image holds image configuration
type Image struct {
	Name        string           `yaml:"name,omitempty" json:",omitempty"`
	Platform    ImagePlatform    `yaml:"platform,omitempty" json:",omitempty"`
	RegOpt      string           `yaml:"regopt,omitempty" json:",omitempty"`
	WatchRepo   bool             `yaml:"watch_repo,omitempty" json:",omitempty"`
	NotifyOn    []NotifyOn       `yaml:"notify_on,omitempty" json:",omitempty"`
	MaxTags     int              `yaml:"max_tags,omitempty" json:",omitempty"`
	SortTags    registry.SortTag `yaml:"sort_tags,omitempty" json:",omitempty"`
	IncludeTags []string         `yaml:"include_tags,omitempty" json:",omitempty"`
	ExcludeTags []string         `yaml:"exclude_tags,omitempty" json:",omitempty"`
	HubTpl      string           `yaml:"hub_tpl,omitempty" json:",omitempty"`
}

// ImagePlatform holds image platform configuration
type ImagePlatform struct {
	OS      string `yaml:"os,omitempty" json:",omitempty"`
	Arch    string `yaml:"arch,omitempty" json:",omitempty"`
	Variant string `yaml:"variant,omitempty" json:",omitempty"`
}

// ImageStatus constants
const (
	ImageStatusNew      = ImageStatus("new")
	ImageStatusUpdate   = ImageStatus("update")
	ImageStatusUnchange = ImageStatus("unchange")
	ImageStatusSkip     = ImageStatus("skip")
	ImageStatusError    = ImageStatus("error")
)

// ImageStatus holds Docker image status analysis
type ImageStatus string

// NotifyOn constants
const (
	NotifyOnNew    = NotifyOn(ImageStatusNew)
	NotifyOnUpdate = NotifyOn(ImageStatusUpdate)
)

// NotifyOn holds notify status type
type NotifyOn string

// NotifyOnDefaults are the default notify status
var NotifyOnDefaults = []NotifyOn{
	NotifyOnNew,
	NotifyOnUpdate,
}

// Valid checks notify status is valid
func (ns *NotifyOn) Valid() bool {
	return ns.OneOf(NotifyOnDefaults)
}

// OneOf checks if notify status is one of the values in the list
func (ns *NotifyOn) OneOf(nsl []NotifyOn) bool {
	for _, n := range nsl {
		if n == *ns {
			return true
		}
	}
	return false
}
