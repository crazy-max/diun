package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path"
	"regexp"

	"github.com/crazy-max/diun/internal/model"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config struct {
	Flags      model.Flags
	App        model.App
	Db         model.Db                  `yaml:"db,omitempty"`
	Watch      model.Watch               `yaml:"watch,omitempty"`
	Notif      model.Notif               `yaml:"notif,omitempty"`
	Registries map[string]model.Registry `yaml:"registries,omitempty"`
	Items      []model.Item              `yaml:"items,omitempty"`
}

// Load returns Configuration struct
func Load(fl model.Flags, version string) (*Config, error) {
	var err error
	var cfg = Config{
		Flags: fl,
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
			Workers:  10,
			Schedule: "0 0 * * * *",
			Os:       "linux",
			Arch:     "amd64",
		},
		Notif: model.Notif{
			Mail: model.Mail{
				Enable:             false,
				Host:               "localhost",
				Port:               25,
				SSL:                false,
				InsecureSkipVerify: false,
			},
			Webhook: model.Webhook{
				Enable:  false,
				Method:  "GET",
				Timeout: 10,
			},
		},
	}

	if _, err = os.Lstat(fl.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(fl.Cfgfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file, %s", err)
	}

	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return &cfg, nil
}

// Check verifies Config values
func (cfg *Config) Check() error {
	if cfg.Flags.Docker {
		cfg.Db.Path = "/data/diun.db"
	}

	if cfg.Db.Path == "" {
		return errors.New("database path is required")
	}
	cfg.Db.Path = path.Clean(cfg.Db.Path)

	for id, reg := range cfg.Registries {
		if err := mergo.Merge(&reg, model.Registry{
			InsecureTLS: false,
			Timeout:     10,
		}); err != nil {
			return fmt.Errorf("cannot set default registry values for %s: %v", id, err)
		}
		cfg.Registries[id] = reg
	}

	for key, item := range cfg.Items {
		if item.Image == "" {
			return fmt.Errorf("image is required for item %d", key)
		}

		if err := mergo.Merge(&item, model.Item{
			WatchRepo: false,
			MaxTags:   25,
		}); err != nil {
			return fmt.Errorf("cannot set default item values for %s: %v", item.Image, err)
		}

		if item.RegistryID != "" {
			reg, found := cfg.Registries[item.RegistryID]
			if !found {
				return fmt.Errorf("registry ID '%s' not found", item.RegistryID)
			}
			cfg.Items[key].Registry = reg
		}

		for _, includeTag := range item.IncludeTags {
			if _, err := regexp.Compile(includeTag); err != nil {
				return fmt.Errorf("include tag regex '%s' for '%s' image cannot compile, %v", item.Image, includeTag, err)
			}
		}

		for _, excludeTag := range item.ExcludeTags {
			if _, err := regexp.Compile(excludeTag); err != nil {
				return fmt.Errorf("exclude tag regex '%s' for '%s' image cannot compile, %v", item.Image, excludeTag, err)
			}
		}

		if err := mergo.Merge(&cfg.Items[key], item); err != nil {
			return err
		}
	}

	if cfg.Notif.Mail.Enable {
		if _, err := mail.ParseAddress(cfg.Notif.Mail.From); err != nil {
			return fmt.Errorf("cannot parse sender mail address, %v", err)
		}
		if _, err := mail.ParseAddress(cfg.Notif.Mail.To); err != nil {
			return fmt.Errorf("cannot parse recipient mail address, %v", err)
		}
	}

	return nil
}

// Display logs configuration in a pretty JSON format
func (cfg *Config) Display() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debug().Msg(string(b))
}
