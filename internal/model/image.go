package model

// RegOpts holds registry options configuration
type RegOpts struct {
	Username    string `yaml:"username,omitempty" json:",omitempty"`
	Password    string `yaml:"password,omitempty" json:",omitempty"`
	InsecureTLS bool   `yaml:"insecure_tls,omitempty" json:",omitempty"`
	Timeout     int    `yaml:"timeout,omitempty" json:",omitempty"`
}

// Image holds image configuration
type Image struct {
	Name            string   `yaml:"name,omitempty" json:",omitempty"`
	Os              string   `yaml:"os,omitempty" json:",omitempty"`
	Arch            string   `yaml:"arch,omitempty" json:",omitempty"`
	RegOptsID       string   `yaml:"regopts_id,omitempty" json:",omitempty"`
	WatchRepo       bool     `yaml:"watch_repo,omitempty" json:",omitempty"`
	MaxTags         int      `yaml:"max_tags,omitempty" json:",omitempty"`
	IncludeTags     []string `yaml:"include_tags,omitempty" json:",omitempty"`
	ExcludeTags     []string `yaml:"exclude_tags,omitempty" json:",omitempty"`
	RegOpts         RegOpts  `yaml:"-" json:"-"`
	SourceContainer string   `yaml:"-" json:"-"`
}
