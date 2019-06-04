package model

// Item holds item configuration for a Docker image
type Item struct {
	Image       string   `yaml:"image,omitempty"`
	RegCredID   string   `yaml:"reg_cred_id,omitempty"`
	InsecureTLS bool     `yaml:"insecure_tls,omitempty"`
	WatchRepo   bool     `yaml:"watch_repo,omitempty"`
	IncludeTags []string `yaml:"include_tags,omitempty"`
	ExcludeTags []string `yaml:"exclude_tags,omitempty"`
	Timeout     int      `yaml:"timeout,omitempty"`
	RegCred     RegCred  `json:"-"`
}
