package model

// RegCred holds registry credential
type RegCred struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}
