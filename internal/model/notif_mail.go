package model

import (
	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifMail holds mail notification configuration details
type NotifMail struct {
	Host               string `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port               int    `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
	SSL                *bool  `yaml:"ssl,omitempty" json:"ssl,omitempty" validate:"required"`
	InsecureSkipVerify *bool  `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty" validate:"required"`
	Username           string `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile       string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password           string `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile       string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	From               string `yaml:"from,omitempty" json:"from,omitempty" validate:"required,email"`
	To                 string `yaml:"to,omitempty" json:"to,omitempty" validate:"required,email"`
}

// GetDefaults gets the default values
func (s *NotifMail) GetDefaults() *NotifMail {
	n := &NotifMail{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifMail) SetDefaults() {
	s.Host = "localhost"
	s.Port = 25
	s.SSL = utl.NewFalse()
	s.InsecureSkipVerify = utl.NewFalse()
}
