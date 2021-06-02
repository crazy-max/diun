package model

// Image holds image configuration
type Image struct {
	Name        string        `yaml:"name,omitempty" json:",omitempty"`
	Platform    ImagePlatform `yaml:"platform,omitempty" json:",omitempty"`
	RegOpt      string        `yaml:"regopt,omitempty" json:",omitempty"`
	WatchRepo   bool          `yaml:"watch_repo,omitempty" json:",omitempty"`
	MaxTags     int           `yaml:"max_tags,omitempty" json:",omitempty"`
	IncludeTags []string      `yaml:"include_tags,omitempty" json:",omitempty"`
	ExcludeTags []string      `yaml:"exclude_tags,omitempty" json:",omitempty"`
	HubTpl      string        `yaml:"hub_tpl,omitempty" json:",omitempty"`
	RepoDigests []string      `yaml:"-" json:"-"`
}

// ImagePlatform holds image platform configuration
type ImagePlatform struct {
	Os      string `yaml:"os,omitempty" json:",omitempty"`
	Arch    string `yaml:"arch,omitempty" json:",omitempty"`
	Variant string `yaml:"variant,omitempty" json:",omitempty"`
}

// Image status constants
const (
	ImageStatusNew      = ImageStatus("new")
	ImageStatusUpdate   = ImageStatus("update")
	ImageStatusUnchange = ImageStatus("unchange")
	ImageStatusSkip     = ImageStatus("skip")
	ImageStatusError    = ImageStatus("error")
)

// ImageStatus holds Docker image status analysis
type ImageStatus string
