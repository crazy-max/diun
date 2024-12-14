package gonfig

import (
	"github.com/crazy-max/gonfig/flag"
	"github.com/pkg/errors"
)

// FlagLoader is the structure representring a flag loader.
type FlagLoader struct {
	//nolint:structcheck,unused
	filename string
	cfg      FlagLoaderConfig
}

// FlagLoaderConfig loads a configuration from flags.
type FlagLoaderConfig struct {
	// Args are command line arguments.
	Args []string
}

// New creates a new Loader from the FlagLoaderConfig cfg.
func NewFlagLoader(cfg FlagLoaderConfig) *FlagLoader {
	return &FlagLoader{
		cfg: cfg,
	}
}

// Load loads the configuration from flags.
func (l *FlagLoader) Load(cfg interface{}) (bool, error) {
	if len(l.cfg.Args) == 0 {
		return false, nil
	}

	if err := flag.Decode(l.cfg.Args, cfg); err != nil {
		return false, errors.Wrap(err, "Failed to decode configuration from flags")
	}

	return true, nil
}
