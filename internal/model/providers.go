package model

// Providers represents a provider configuration
type Providers struct {
	Docker *PrdDocker `yaml:"docker,omitempty" json:",omitempty"`
	Swarm  *PrdSwarm  `yaml:"swarm,omitempty" json:",omitempty"`
	File   *PrdFile   `yaml:"file,omitempty" json:",omitempty"`
}

// PrdDocker holds docker provider configuration
type PrdDocker struct {
	Endpoint       string `yaml:"endpoint,omitempty" json:",omitempty"`
	APIVersion     string `yaml:"api_version,omitempty" json:",omitempty"`
	TLSCertsPath   string `yaml:"tls_certs_path,omitempty" json:",omitempty"`
	TLSVerify      *bool  `yaml:"tls_verify,omitempty" json:",omitempty"`
	WatchByDefault *bool  `yaml:"watch_by_default,omitempty" json:",omitempty"`
	WatchStopped   *bool  `yaml:"watch_stopped,omitempty" json:",omitempty"`
}

// PrdSwarm holds swarm provider configuration
type PrdSwarm struct {
	Endpoint       string `yaml:"endpoint,omitempty" json:",omitempty"`
	APIVersion     string `yaml:"api_version,omitempty" json:",omitempty"`
	TLSCertsPath   string `yaml:"tls_certs_path,omitempty" json:",omitempty"`
	TLSVerify      *bool  `yaml:"tls_verify,omitempty" json:",omitempty"`
	WatchByDefault *bool  `yaml:"watch_by_default,omitempty" json:",omitempty"`
}

// PrdFile holds file provider configuration
type PrdFile struct {
	Filename  string `yaml:"filename,omitempty" json:",omitempty"`
	Directory string `yaml:"directory,omitempty" json:",omitempty"`
}
