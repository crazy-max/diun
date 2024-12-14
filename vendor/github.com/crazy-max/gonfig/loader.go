package gonfig

// Loader is a configuration resource loader.
type Loader interface {
	// Load populates cfg.
	Load(cfg interface{}) (bool, error)
}
