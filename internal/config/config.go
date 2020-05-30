package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/pkg/traefik/config/env"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config model.Config

// Load returns Configuration struct
func Load(cli model.Cli, version string) (*Config, error) {
	var cfg = Config{
		Cli: cli,
		App: model.App{
			Version: version,
		},
	}

	if err := mergo.Merge(&cfg, Config(model.DefaultConfig)); err != nil {
		return nil, errors.Wrap(err, "cannot set default values for config")
	}

	if err := cfg.loadFile(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.loadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) loadFile(out interface{}) error {
	if _, err := os.Lstat(cfg.Cli.Cfgfile); os.IsNotExist(err) {
		log.Debug().Msg("No config file provided")
		return nil
	}

	b, err := ioutil.ReadFile(cfg.Cli.Cfgfile)
	if err != nil {
		return errors.Wrap(err, "unable to read config file")
	}

	if err := yaml.UnmarshalStrict(b, out); err != nil {
		return errors.Wrap(err, "unable to decode into struct")
	}

	return nil
}

func (cfg *Config) loadEnv(out interface{}) error {
	var envvars []string
	for _, envvar := range env.FindPrefixedEnvVars(os.Environ(), "DIUN_", out) {
		if strings.HasPrefix(envvar, "DIUN_APP") || strings.HasPrefix(envvar, "DIUN_CLI") {
			continue
		}
		envvars = append(envvars, envvar)
	}
	if len(envvars) == 0 {
		return nil
	}

	if err := env.Decode(envvars, "DIUN_", out); err != nil {
		log.Debug().Strs("envvars", envvars).Msg("Environment variables")
		return errors.Wrap(err, "failed to decode configuration from environment variables")
	}

	return nil
}

// Display configuration in a pretty JSON format
func (cfg *Config) Display() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}
