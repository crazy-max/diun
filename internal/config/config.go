package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/containous/traefik/v2/pkg/config/env"
	"github.com/containous/traefik/v2/pkg/config/file"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Config holds configuration details
type Config struct {
	Db        *model.Db        `yaml:"db,omitempty" json:"db,omitempty"`
	Watch     *model.Watch     `yaml:"watch,omitempty" json:"watch,omitempty"`
	Notif     *model.Notif     `yaml:"notif,omitempty" json:"notif,omitempty"`
	RegOpts   model.RegOpts    `yaml:"regopts,omitempty" json:"regopts,omitempty" validate:"unique=Name,dive"`
	Providers *model.Providers `yaml:"providers,omitempty" json:"providers,omitempty" validate:"required"`
}

// Load returns Config struct
func Load(cfgfile string) (*Config, error) {
	cfg := Config{
		Db:    (&model.Db{}).GetDefaults(),
		Watch: (&model.Watch{}).GetDefaults(),
	}

	if err := cfg.loadFile(cfgfile, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.loadEnv(&cfg); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) loadFile(cfgfile string, out interface{}) error {
	if len(cfgfile) == 0 {
		log.Debug().Msg("No configuration file defined")
		return nil
	}

	if _, err := os.Lstat(cfgfile); os.IsNotExist(err) {
		return fmt.Errorf("config file %s not found", cfgfile)
	}

	if err := file.Decode(cfgfile, out); err != nil {
		return errors.Wrap(err, "failed to decode configuration from file")
	}

	return nil
}

func (cfg *Config) loadEnv(out interface{}) error {
	var envvars []string
	for _, envvar := range env.FindPrefixedEnvVars(os.Environ(), "DIUN_", out) {
		envvars = append(envvars, envvar)
	}
	if len(envvars) == 0 {
		log.Debug().Msg("No DIUN_* environment variables defined")
		return nil
	}

	if err := env.Decode(envvars, "DIUN_", out); err != nil {
		return errors.Wrap(err, "failed to decode configuration from environment variables")
	}

	return nil
}

// String returns the string representation of configuration
func (cfg *Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}
