package model

import (
	"strings"

	"github.com/crazy-max/diun/v3/pkg/utl"
)

// Config holds configuration details
type Config struct {
	Cli       Cli                `yaml:"-"`
	App       App                `yaml:"-"`
	Db        Db                 `yaml:"db,omitempty"`
	Watch     Watch              `yaml:"watch,omitempty"`
	Notif     *Notif             `yaml:"notif,omitempty"`
	RegOpts   map[string]RegOpts `yaml:"regopts,omitempty"`
	Providers *Providers         `yaml:"providers,omitempty"`
}

// DefaultConfig holds default configuration
var DefaultConfig = Config{
	App: App{
		ID:     "diun",
		Name:   "Diun",
		Desc:   "Docker image update notifier",
		URL:    "https://github.com/crazy-max/diun",
		Author: "CrazyMax",
	},
	Db: Db{
		Path: "diun.db",
	},
	Watch: Watch{
		Workers:         10,
		Schedule:        "0 * * * *",
		FirstCheckNotif: utl.NewFalse(),
	},
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if s.RegOpts != nil {
		for id, regopt := range s.RegOpts {
			delete(s.RegOpts, id)
			s.RegOpts[strings.ToLower(id)] = regopt
		}
	}

	return nil
}
