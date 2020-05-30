package model

import (
	"time"

	"github.com/crazy-max/diun/v3/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// RegOpts holds registry options configuration
type RegOpts struct {
	Username     string         `yaml:"username,omitempty" json:",omitempty"`
	UsernameFile string         `yaml:"username_file,omitempty" json:",omitempty"`
	Password     string         `yaml:"password,omitempty" json:",omitempty"`
	PasswordFile string         `yaml:"password_file,omitempty" json:",omitempty"`
	InsecureTLS  *bool          `yaml:"insecure_tls,omitempty" json:",omitempty"`
	Timeout      *time.Duration `yaml:"timeout,omitempty" json:",omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *RegOpts) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RegOpts
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, RegOpts{
		InsecureTLS: utl.NewFalse(),
		Timeout:     utl.NewDuration(10 * time.Second),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for registry options")
	}

	return nil
}
