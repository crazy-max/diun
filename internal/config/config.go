package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config struct {
	Cli       model.Cli
	App       model.App
	Db        model.Db                 `yaml:"db,omitempty"`
	Watch     model.Watch              `yaml:"watch,omitempty"`
	Notif     model.Notif              `yaml:"notif,omitempty"`
	RegOpts   map[string]model.RegOpts `yaml:"regopts,omitempty"`
	Providers model.Providers          `yaml:"providers,omitempty"`
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
			FirstCheckNotif: false,
		},
		Notif: model.Notif{
			Gotify: model.NotifGotify{
				Enable:  false,
				Timeout: 10,
			},
			Mail: model.NotifMail{
				Enable:             false,
				Host:               "localhost",
				Port:               25,
				SSL:                false,
				InsecureSkipVerify: false,
			},
			RocketChat: model.NotifRocketChat{
				Enable:  false,
				Timeout: 10,
			},
			Slack: model.NotifSlack{
				Enable: false,
			},
			Telegram: model.NotifTelegram{
				Enable: false,
			},
			Webhook: model.NotifWebhook{
				Enable:  false,
				Method:  "GET",
				Timeout: 10,
			},
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

	for id, regopts := range cfg.RegOpts {
		if err := cfg.validateRegOpts(id, regopts); err != nil {
			return err
		}
	}

	for id, prdDocker := range cfg.Providers.Docker {
		if err := cfg.validateDockerProvider(id, prdDocker); err != nil {
			return err
		}
	}

	for id, prdSwarm := range cfg.Providers.Swarm {
		if err := cfg.validateSwarmProvider(id, prdSwarm); err != nil {
			return err
		}
	}

	if err := cfg.validateFileProvider(); err != nil {
		return err
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
		return fmt.Errorf("cannot set default values for registry options %s: %v", id, err)
	}

	cfg.RegOpts[id] = regopts
	return nil
}

func (cfg *Config) validateDockerProvider(id string, prdDocker model.PrdDocker) error {
	if err := mergo.Merge(&prdDocker, model.PrdDocker{
		TLSVerify:      true,
		WatchByDefault: false,
		WatchStopped:   false,
	}); err != nil {
		return fmt.Errorf("cannot set default values for docker %s provider: %v", id, err)
	}

	cfg.Providers.Docker[id] = prdDocker
	return nil
}

func (cfg *Config) validateSwarmProvider(id string, prdSwarm model.PrdSwarm) error {
	if err := mergo.Merge(&prdSwarm, model.PrdSwarm{
		TLSVerify:      true,
		WatchByDefault: false,
	}); err != nil {
		return fmt.Errorf("cannot set default values for swarm %s provider: %v", id, err)
	}

	cfg.Providers.Swarm[id] = prdSwarm
	return nil
}

func (cfg *Config) validateFileProvider() error {
	switch {
	case len(cfg.Providers.File.Directory) > 0:
		if _, err := os.Stat(cfg.Providers.File.Directory); os.IsNotExist(err) {
			return errors.Wrap(err, "directory not found for file provider")
		}
	case len(cfg.Providers.File.Filename) > 0:
		if _, err := os.Stat(cfg.Providers.File.Filename); os.IsNotExist(err) {
			return errors.Wrap(err, "filename not found for file provider")
		}
	}
	return nil
}

// Display configuration in a pretty JSON format
func (cfg *Config) Display() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debug().Msg(string(b))
}
