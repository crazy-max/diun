package model

// Image holds image configuration
type Image struct {
	Name        string   `yaml:"name,omitempty" json:",omitempty"`
	Os          string   `yaml:"os,omitempty" json:",omitempty"`
	Arch        string   `yaml:"arch,omitempty" json:",omitempty"`
	RegOptsID   string   `yaml:"regopts_id,omitempty" json:",omitempty"`
	WatchRepo   bool     `yaml:"watch_repo,omitempty" json:",omitempty"`
	MaxTags     int      `yaml:"max_tags,omitempty" json:",omitempty"`
	IncludeTags []string `yaml:"include_tags,omitempty" json:",omitempty"`
	ExcludeTags []string `yaml:"exclude_tags,omitempty" json:",omitempty"`
}

// Image status constants
const (
	ImageStatusNew      = ImageStatus("new")
	ImageStatusUpdate   = ImageStatus("update")
	ImageStatusUnchange = ImageStatus("unchange")
)

// ImageStatus holds Docker image status analysis
type ImageStatus string
