package model

import (
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// Watch holds data necessary for watch configuration
type Watch struct {
	Workers         int    `yaml:"workers,omitempty"`
	Schedule        string `yaml:"schedule,omitempty"`
	FirstCheckNotif *bool  `yaml:"first_check_notif,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *Watch) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Watch
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, DefaultConfig.Watch); err != nil {
		return errors.Wrap(err, "cannot set default values for watch")
	}

	return nil
}
