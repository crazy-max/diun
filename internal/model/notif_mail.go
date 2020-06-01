package model

import (
	"net/mail"

	"github.com/crazy-max/diun/v3/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
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

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifMail) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifMail
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(s.From); err != nil {
		return errors.Wrap(err, "cannot parse sender mail address")
	}
	if _, err := mail.ParseAddress(s.To); err != nil {
		return errors.Wrap(err, "cannot parse recipient mail address")
	}

	if err := mergo.Merge(s, NotifMail{
		Host:               "localhost",
		Port:               25,
		SSL:                utl.NewFalse(),
		InsecureSkipVerify: utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for mail notif")
	}

	return nil
}
