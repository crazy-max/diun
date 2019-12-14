package model

// Providers represents a provider configuration
type Providers struct {
	Docker []PrdDocker `yaml:"docker,omitempty" json:",omitempty"`
	Swarm  []PrdSwarm  `yaml:"swarm,omitempty" json:",omitempty"`
	Static []PrdStatic `yaml:"static,omitempty" json:",omitempty"`
}

// PrdDocker holds docker provider configuration
type PrdDocker struct {
	ID             string `yaml:"id,omitempty" json:",omitempty"`
	Endpoint       string `yaml:"endpoint,omitempty" json:",omitempty"`
	ApiVersion     string `yaml:"api_version,omitempty" json:",omitempty"`
	TLSCertsPath   string `yaml:"tls_certs_path,omitempty" json:",omitempty"`
	TLSVerify      bool   `yaml:"tls_verify,omitempty" json:",omitempty"`
	WatchByDefault bool   `yaml:"watch_by_default,omitempty" json:",omitempty"`
	WatchStopped   bool   `yaml:"watch_stopped,omitempty" json:",omitempty"`
}

// PrdSwarm holds swarm provider configuration
type PrdSwarm struct {
	ID             string `yaml:"id,omitempty" json:",omitempty"`
	Endpoint       string `yaml:"endpoint,omitempty" json:",omitempty"`
	ApiVersion     string `yaml:"api_version,omitempty" json:",omitempty"`
	TLSCertsPath   string `yaml:"tls_certs_path,omitempty" json:",omitempty"`
	TLSVerify      bool   `yaml:"tls_verify,omitempty" json:",omitempty"`
	WatchByDefault bool   `yaml:"watch_by_default,omitempty" json:",omitempty"`
}

// PrdStatic holds static provider configuration
type PrdStatic Image
