package model

// PrdFile holds file provider configuration
type PrdFile struct {
	Filename  string `yaml:"filename,omitempty" json:"filename,omitempty" validate:"omitempty,file"`
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty" validate:"omitempty,dir"`
}

// GetDefaults gets the default values
func (s *PrdFile) GetDefaults() *PrdFile {
	return nil
}

// SetDefaults sets the default values
func (s *PrdFile) SetDefaults() {
	// noop
}
