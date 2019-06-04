package model

// Db holds data necessary for database configuration
type Db struct {
	Path string `yaml:"path,omitempty"`
}
