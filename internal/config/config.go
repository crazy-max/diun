package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/utl"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config holds configuration details
type Config struct {
	Flags   model.Flags
	App     model.App
	Db      model.Db                 `yaml:"db,omitempty"`
	Watch   model.Watch              `yaml:"watch,omitempty"`
	Notif   model.Notif              `yaml:"notif,omitempty"`
	RegOpts map[string]model.RegOpts `yaml:"regopts,omitempty"`
	Image   []model.Image            `yaml:"image,omitempty"`
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
			Schedule: "0 0 * * * *",
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

	if _, err = os.Lstat(flags.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	cfgBytes, err := ioutil.ReadFile(flags.Cfgfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file, %s", err)
	}

	// Replace environment variables in configuration file
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}
		bytes.ReplaceAll(cfgBytes, []byte("${" + parts[0] + "}"), []byte(value))
	}

	if err := yaml.UnmarshalStrict(cfgBytes, &cfg); err != nil {
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
		if regopts, err := cfg.validateRegOpts(id, regopts); err != nil {
			return err
		} else {
			cfg.RegOpts[id] = regopts
		}
	}

	for key, img := range cfg.Image {
		if img, err := cfg.validateImage(key, img); err != nil {
			return err
		} else {
			cfg.Image[key] = img
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

func (cfg *Config) validateRegOpts(id string, regopts model.RegOpts) (model.RegOpts, error) {
	defTimeout := 10
	if regopts.Timeout <= 0 {
		defTimeout = 0
	}

	if err := mergo.Merge(&regopts, model.RegOpts{
		InsecureTLS: false,
		Timeout:     defTimeout,
	}); err != nil {
		return model.RegOpts{}, fmt.Errorf("cannot set default registry options values for %s: %v", id, err)
	}

	cfg.RegOpts[id] = regopts
	return model.RegOpts{}, nil
}

func (cfg *Config) PutRegOpts(id string, regopts model.RegOpts) error {
	regopts, err := cfg.validateRegOpts(id, regopts)
	if err != nil {
		return err
	}
	cfg.RegOpts[id] = regopts
	return nil
}
func (cfg *Config) RemoveRegOpts(id string) error {
	if _, ok := cfg.RegOpts[id]; !ok {
		return errors.New("no entry with this id")
	}
	delete(cfg.RegOpts, id)
	return nil
}

func (cfg *Config) validateImage(key int, img model.Image) (model.Image, error) {
	if img.Name == "" {
		return model.Image{}, fmt.Errorf("name is required for image %d", key)
	}

	if err := mergo.Merge(&img, model.Image{
		Os:        "linux",
		Arch:      "amd64",
		WatchRepo: false,
		MaxTags:   0,
	}); err != nil {
		return model.Image{}, fmt.Errorf("cannot set default image values for %s: %v", img.Name, err)
	}

	if img.RegOptsID != "" {
		regopts, found := cfg.RegOpts[img.RegOptsID]
		if !found {
			return model.Image{}, fmt.Errorf("registry options %s not found for %s", img.RegOptsID, img.Name)
		}
		img.RegOpts = regopts
	}

	for _, includeTag := range img.IncludeTags {
		if _, err := regexp.Compile(includeTag); err != nil {
			return model.Image{}, fmt.Errorf("include tag regex '%s' for %s cannot compile, %v", includeTag, img.Name, err)
		}
	}

	for _, excludeTag := range img.ExcludeTags {
		if _, err := regexp.Compile(excludeTag); err != nil {
			return model.Image{}, fmt.Errorf("exclude tag regex '%s' for '%s' image cannot compile, %v", img.Name, excludeTag, err)
		}
	}

	return img, nil
}

func (cfg *Config) AddImage(img model.Image) error {
	img, err := cfg.validateImage(len(cfg.Image), img)
	if err != nil {
		return err
	}
	cfg.Image = append(cfg.Image, img)
	return nil
}
func (cfg *Config) SetImage(key int, img model.Image) error {
	img, err := cfg.validateImage(key, img)
	if err != nil {
		return err
	}
	cfg.Image[key] = img
	return nil
}
func (cfg *Config) RemoveImage(key int) error {
	if key < 0 || key >= len(cfg.Image) {
		return errors.New("index out of range")
	}
	cfg.Image = append(cfg.Image[:key], cfg.Image[key+1:]...)
	return nil
}

// Display logs configuration in a pretty JSON format
func (cfg *Config) Display() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debug().Msg(string(b))
}
