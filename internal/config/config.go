package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config struct {
	Cli       model.Cli
	App       model.App
	Db        model.Db                 `yaml:"db,omitempty"`
	Watch     model.Watch              `yaml:"watch,omitempty"`
	Notif     *model.Notif             `yaml:"notif,omitempty"`
	RegOpts   map[string]model.RegOpts `yaml:"regopts,omitempty"`
	Providers *model.Providers         `yaml:"providers,omitempty"`
}

// Load returns Configuration struct
func Load(cli model.Cli, version string) (*Config, error) {
	var err error
	var cfg = Config{
		Cli: cli,
		App: model.App{
			ID:      "diun",
			Name:    "Diun",
			Desc:    "Docker image update notifier",
			URL:     "https://github.com/crazy-max/diun",
			Author:  "CrazyMax",
			Version: version,
		},
		Db: model.Db{
			Path: "diun.db",
		},
		Watch: model.Watch{
			Workers:         10,
			Schedule:        "0 * * * *",
			FirstCheckNotif: utl.NewFalse(),
		},
	}

	if _, err = os.Lstat(cli.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(cli.Cfgfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file, %s", err)
	}

	if err := yaml.UnmarshalStrict(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	cfg.Db.Path = utl.GetEnv("DIUN_DB", cfg.Db.Path)
	if cfg.Db.Path == "" {
		return errors.New("database path is required")
	}
	cfg.Db.Path = path.Clean(cfg.Db.Path)

	if err := cfg.validateNotif(); err != nil {
		return err
	}
	if err := cfg.validateRegopts(); err != nil {
		return err
	}
	if err := cfg.validateProviders(); err != nil {
		return err
	}

	return nil
}

// Display configuration in a pretty JSON format
func (cfg *Config) Display() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}
