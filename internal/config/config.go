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
	Flags    model.Flags
	App      model.App
	Db       model.Db                 `yaml:"db,omitempty"`
	Watch    model.Watch              `yaml:"watch,omitempty"`
	Notif    model.Notif              `yaml:"notif,omitempty"`
	RegCreds map[string]model.RegCred `yaml:"reg_creds,omitempty"`
	Items    []model.Item             `yaml:"items,omitempty"`
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
			Schedule: "0 */30 * * * *",
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

	for id, regCred := range cfg.RegCreds {
		if regCred.Username == "" || regCred.Password == "" {
			return fmt.Errorf("username and password required for registry credentials '%s'", id)
		}
	}

	for key, item := range cfg.Items {
		if item.RegCredID != "" {
			regCred, found := cfg.RegCreds[item.RegCredID]
			if !found {
				return fmt.Errorf("registry credentials '%s' not found", item.RegCredID)
			}
			cfg.Items[key].RegCred = regCred
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

		if err := mergo.Merge(&cfg.Items[key], model.Item{
			Timeout: 5,
		}); err != nil {
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
