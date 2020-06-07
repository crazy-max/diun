package model

// Db holds data necessary for database configuration
type Db struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *Db) GetDefaults() *Db {
	n := &Db{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *Db) SetDefaults() {
	s.Path = "diun.db"
}
