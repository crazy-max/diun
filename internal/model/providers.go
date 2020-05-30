package model

import (
	"os"

	"github.com/crazy-max/diun/v3/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// Providers represents a provider configuration
type Providers struct {
	Docker *PrdDocker `yaml:"docker,omitempty" json:",omitempty"`
	Swarm  *PrdSwarm  `yaml:"swarm,omitempty" json:",omitempty"`
	File   *PrdFile   `yaml:"file,omitempty" json:",omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *Providers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Providers
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
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

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *PrdDocker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain PrdDocker
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, PrdDocker{
		TLSVerify:      utl.NewTrue(),
		WatchByDefault: utl.NewFalse(),
		WatchStopped:   utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for docker provider")
	}

	return nil
}

// PrdSwarm holds swarm provider configuration
type PrdSwarm struct {
	Endpoint       string `yaml:"endpoint,omitempty" json:",omitempty"`
	APIVersion     string `yaml:"api_version,omitempty" json:",omitempty"`
	TLSCertsPath   string `yaml:"tls_certs_path,omitempty" json:",omitempty"`
	TLSVerify      *bool  `yaml:"tls_verify,omitempty" json:",omitempty"`
	WatchByDefault *bool  `yaml:"watch_by_default,omitempty" json:",omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *PrdSwarm) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain PrdSwarm
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, PrdSwarm{
		TLSVerify:      utl.NewTrue(),
		WatchByDefault: utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for docker provider")
	}

	return nil
}

// PrdFile holds file provider configuration
type PrdFile struct {
	Filename  string `yaml:"filename,omitempty" json:",omitempty"`
	Directory string `yaml:"directory,omitempty" json:",omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *PrdFile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain PrdFile
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	switch {
	case len(s.Directory) > 0:
		if _, err := os.Stat(s.Directory); os.IsNotExist(err) {
			return errors.Wrap(err, "directory not found for file provider")
		}
	case len(s.Filename) > 0:
		if _, err := os.Stat(s.Filename); os.IsNotExist(err) {
			return errors.Wrap(err, "filename not found for file provider")
		}
	default:
		return errors.New("error using file provider, neither filename or directory defined")
	}

	return nil
}
