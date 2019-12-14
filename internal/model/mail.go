package model

// Mail holds mail notification configuration details
type Mail struct {
	Enable             bool   `yaml:"enable,omitempty"`
	Host               string `yaml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	SSL                bool   `yaml:"ssl,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
	Username           string `yaml:"username,omitempty"`
	UsernameFile       string `yaml:"username_file,omitempty"`
	Password           string `yaml:"password,omitempty"`
	PasswordFile       string `yaml:"password_file,omitempty"`
	From               string `yaml:"from,omitempty"`
	To                 string `yaml:"to,omitempty"`
}
