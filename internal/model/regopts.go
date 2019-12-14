package model

// RegOpts holds registry options configuration
type RegOpts struct {
	Username     string `yaml:"username,omitempty" json:",omitempty"`
	UsernameFile string `yaml:"username_file,omitempty" json:",omitempty"`
	Password     string `yaml:"password,omitempty" json:",omitempty"`
	PasswordFile string `yaml:"password_file,omitempty" json:",omitempty"`
	InsecureTLS  bool   `yaml:"insecure_tls,omitempty" json:",omitempty"`
	Timeout      int    `yaml:"timeout,omitempty" json:",omitempty"`
}
