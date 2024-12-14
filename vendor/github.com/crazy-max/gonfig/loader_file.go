package gonfig

import (
	"github.com/crazy-max/gonfig/file"
)

// FileLoader is the structure representring a file loader.
type FileLoader struct {
	filename string
	cfg      FileLoaderConfig
}

// FileLoader loads a configuration from a file.
type FileLoaderConfig struct {
	Filename string
	Finder   Finder
}

// New creates a new Loader fromt the FileLoaderConfig cfg.
func NewFileLoader(cfg FileLoaderConfig) *FileLoader {
	return &FileLoader{
		cfg: cfg,
	}
}

// GetFilename returns the configuration file if any.
func (l *FileLoader) GetFilename() string {
	return l.filename
}

// Load loads the configuration from a file and/or finders.
func (l *FileLoader) Load(cfg interface{}) (bool, error) {
	var err error

	l.filename, err = l.cfg.Finder.Find(l.cfg.Filename)
	if err != nil {
		return false, err
	}

	if len(l.filename) == 0 {
		return false, nil
	}

	if err = file.Decode(l.filename, cfg); err != nil {
		return false, err
	}

	return true, nil
}
