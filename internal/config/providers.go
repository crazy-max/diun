package config

import (
	"os"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

func (cfg *Config) validateProviders() error {
	if cfg.Providers == nil || (cfg.Providers.Docker == nil && cfg.Providers.Swarm == nil && cfg.Providers.File == nil) {
		return errors.New("At least one provider is required")
	}

	if err := cfg.validateProviderDocker(); err != nil {
		return err
	}
	if err := cfg.validateProviderSwarm(); err != nil {
		return err
	}
	if err := cfg.validateProviderFile(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) validateProviderDocker() error {
	if cfg.Providers.Docker == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Providers.Docker, model.PrdDocker{
		TLSVerify:      utl.NewTrue(),
		WatchByDefault: utl.NewFalse(),
		WatchStopped:   utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for docker provider")
	}

	return nil
}

func (cfg *Config) validateProviderSwarm() error {
	if cfg.Providers.Swarm == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Providers.Swarm, model.PrdSwarm{
		TLSVerify:      utl.NewTrue(),
		WatchByDefault: utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for docker provider")
	}

	return nil
}

func (cfg *Config) validateProviderFile() error {
	if cfg.Providers.File == nil {
		return nil
	}

	switch {
	case len(cfg.Providers.File.Directory) > 0:
		if _, err := os.Stat(cfg.Providers.File.Directory); os.IsNotExist(err) {
			return errors.Wrap(err, "directory not found for file provider")
		}
	case len(cfg.Providers.File.Filename) > 0:
		if _, err := os.Stat(cfg.Providers.File.Filename); os.IsNotExist(err) {
			return errors.Wrap(err, "filename not found for file provider")
		}
	default:
		return errors.New("error using file provider, neither filename or directory defined")
	}

	return nil
}
