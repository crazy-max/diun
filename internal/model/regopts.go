package model

import (
	"time"

	"github.com/crazy-max/diun/v3/pkg/utl"
)

// RegOpts holds registry options configuration
type RegOpts struct {
	Username     string         `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile string         `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password     string         `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile string         `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	InsecureTLS  *bool          `yaml:"insecureTls,omitempty" json:"insecureTls,omitempty" validate:"required"`
	Timeout      *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *RegOpts) GetDefaults() *RegOpts {
	n := &RegOpts{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *RegOpts) SetDefaults() {
	s.InsecureTLS = utl.NewFalse()
	s.Timeout = utl.NewDuration(10 * time.Second)
}
