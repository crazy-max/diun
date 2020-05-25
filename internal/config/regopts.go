package config

import (
	"fmt"

	"github.com/crazy-max/diun/internal/model"
	"github.com/imdario/mergo"
)

func (cfg *Config) validateRegopts() error {
	for id, regopt := range cfg.RegOpts {
		if err := cfg.validateRegOpt(id, regopt); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) validateRegOpt(id string, regopts model.RegOpts) error {
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
