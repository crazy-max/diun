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
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config struct {
	Flags     model.Flags
	App       model.App
	Db        model.Db                 `yaml:"db,omitempty"`
	Watch     model.Watch              `yaml:"watch,omitempty"`
	Notif     model.Notif              `yaml:"notif,omitempty"`
	RegOpts   map[string]model.RegOpts `yaml:"regopts,omitempty"`
	Providers model.Providers          `yaml:"providers,omitempty"`
}

// Load returns Configuration struct
func Load(flags model.Flags, version string) (*Config, error) {
	var err error
	var cfg = Config{
		Flags: flags,
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
			Schedule: "0 * * * *",
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
		Providers: model.Providers{
			Docker: []model.PrdDocker{},
			Image:  []model.PrdImage{},
		},
	}

	if _, err = os.Lstat(flags.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(flags.Cfgfile)
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

	for id, regopts := range cfg.RegOpts {
		if err := cfg.validateRegOpts(id, regopts); err != nil {
			return err
		}
	}

	for key, dock := range cfg.Providers.Docker {
		if err := cfg.validateDockerProvider(key, dock); err != nil {
			return err
		}
	}

	for key, img := range cfg.Providers.Image {
		if err := cfg.validateImageProvider(key, img); err != nil {
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

func (cfg *Config) validateRegOpts(id string, regopts model.RegOpts) error {
	defTimeout := 10
	if regopts.Timeout <= 0 {
		defTimeout = 0
	}

	if err := mergo.Merge(&regopts, model.RegOpts{
		InsecureTLS: false,
		Timeout:     defTimeout,
	}); err != nil {
		return fmt.Errorf("cannot set default registry options values for %s: %v", id, err)
	}

	cfg.RegOpts[id] = regopts
	return nil
}

func (cfg *Config) validateDockerProvider(key int, dock model.PrdDocker) error {
	if dock.ID == "" {
		return fmt.Errorf("ID is required for docker provider %d", key)
	}

	if err := mergo.Merge(&dock, model.PrdDocker{
		TLSVerify:      true,
		SwarmMode:      false,
		WatchByDefault: false,
		WatchStopped:   false,
	}); err != nil {
		return fmt.Errorf("cannot set default docker provider values for %s: %v", dock.ID, err)
	}

	cfg.Providers.Docker[key] = dock
	return nil
}

func (cfg *Config) validateImageProvider(key int, img model.PrdImage) error {
	if img.Name == "" {
		return fmt.Errorf("name is required for image provider %d", key)
	}

	if err := mergo.Merge(&img, model.PrdImage{
		Os:        "linux",
		Arch:      "amd64",
		WatchRepo: false,
		MaxTags:   0,
	}); err != nil {
		return fmt.Errorf("cannot set default image image values for %s: %v", img.Name, err)
	}

	for _, includeTag := range img.IncludeTags {
		if _, err := regexp.Compile(includeTag); err != nil {
			return fmt.Errorf("include tag regex '%s' for %s cannot compile, %v", includeTag, img.Name, err)
		}
	}

	for _, excludeTag := range img.ExcludeTags {
		if _, err := regexp.Compile(excludeTag); err != nil {
			return fmt.Errorf("exclude tag regex '%s' for '%s' image cannot compile, %v", img.Name, excludeTag, err)
		}
	}

	cfg.Providers.Image[key] = img
	return nil
}

// Display logs configuration in a pretty JSON format
func (cfg *Config) Display() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debug().Msg(string(b))
}
