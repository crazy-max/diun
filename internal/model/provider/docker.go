package provider

// Docker holds docker provider configuration
type Docker struct {
	ID           string `yaml:"id,omitempty" json:",omitempty"`
	Endpoint     string `yaml:"endpoint,omitempty" json:",omitempty"`
	ApiVersion   string `yaml:"api_version,omitempty" json:",omitempty"`
	CAFile       string `yaml:"ca_file,omitempty" json:",omitempty"`
	CertFile     string `yaml:"cert_file,omitempty" json:",omitempty"`
	KeyFile      string `yaml:"key_file,omitempty" json:",omitempty"`
	TLSVerify    string `yaml:"tls_verify,omitempty" json:",omitempty"`
	SwarmMode    bool   `yaml:"swarm_mode,omitempty" json:",omitempty"`
	WatchStopped bool   `yaml:"watch_stopped,omitempty" json:",omitempty"`
}
