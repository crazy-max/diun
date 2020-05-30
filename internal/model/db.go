package model

import (
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// Db holds data necessary for database configuration
type Db struct {
	Path string `yaml:"path,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *Db) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Db
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, DefaultConfig.Db); err != nil {
		return errors.Wrap(err, "cannot set default values for db")
	}

	return nil
}
