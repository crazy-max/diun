package gonfig

import (
	"os"

	"github.com/crazy-max/gonfig/env"
	"github.com/pkg/errors"
)

// EnvLoader is the structure representring an environment variable loader.
type EnvLoader struct {
	vars []string
	cfg  EnvLoaderConfig
}

// EnvLoaderConfig loads a configuration from environment variables.
type EnvLoaderConfig struct {
	// Prefix to use. Default to "GONFIG_"
	Prefix string
}

// New creates a new Loader from the EnvLoaderConfig cfg.
func NewEnvLoader(cfg EnvLoaderConfig) *EnvLoader {
	return &EnvLoader{
		cfg: cfg,
	}
}

// GetVars returns the environment variables found.
func (l *EnvLoader) GetVars() []string {
	return l.vars
}

// Load loads the configuration from the environment variables.
func (l *EnvLoader) Load(cfg interface{}) (bool, error) {
	prefix := l.cfg.Prefix
	if prefix == "" {
		prefix = env.DefaultNamePrefix
	}

	l.vars = env.FindPrefixedEnvVars(os.Environ(), prefix, cfg)
	if len(l.vars) == 0 {
		return false, nil
	}

	if err := env.Decode(l.vars, prefix, cfg); err != nil {
		return false, errors.Wrap(err, "Failed to decode configuration from environment variables")
	}

	return true, nil
}
