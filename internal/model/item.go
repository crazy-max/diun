package model

// Item holds item configuration for a Docker image
type Item struct {
	Image       string   `yaml:"image,omitempty" json:",omitempty"`
	RegistryID  string   `yaml:"registry_id,omitempty" json:",omitempty"`
	WatchRepo   bool     `yaml:"watch_repo,omitempty" json:",omitempty"`
	MaxTags     int      `yaml:"max_tags,omitempty" json:",omitempty"`
	IncludeTags []string `yaml:"include_tags,omitempty" json:",omitempty"`
	ExcludeTags []string `yaml:"exclude_tags,omitempty" json:",omitempty"`
	Registry    Registry `yaml:"-" json:"-"`
}
