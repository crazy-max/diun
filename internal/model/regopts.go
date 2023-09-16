package model

import (
	"strings"
	"time"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// RegOpts holds slice of registry options
type RegOpts []RegOpt

// RegOpt holds registry options configuration
type RegOpt struct {
	Name         string         `yaml:"name,omitempty" json:"name,omitempty" validate:"required"`
	Selector     RegOptSelector `yaml:"selector,omitempty" json:"selector,omitempty" validate:"required,oneof=name image"`
	Username     string         `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile string         `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password     string         `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile string         `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	InsecureTLS  *bool          `yaml:"insecureTLS,omitempty" json:"insecureTLS,omitempty" validate:"required"`
	Timeout      *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// RegOpt selector constants
const (
	RegOptSelectorName  = RegOptSelector("name")
	RegOptSelectorImage = RegOptSelector("image")
)

// RegOptSelector holds registry options selector
type RegOptSelector string

// GetDefaults gets the default values
func (s *RegOpt) GetDefaults() *RegOpt {
	n := &RegOpt{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *RegOpt) SetDefaults() {
	s.Selector = RegOptSelectorName
	s.InsecureTLS = utl.NewFalse()
	s.Timeout = utl.NewDuration(0)
}

// Select returns a registry based on its selector
func (s *RegOpts) Select(name string, image registry.Image) (*RegOpt, error) {
	for _, regOpt := range *s {
		if regOpt.Selector == RegOptSelectorName && name == regOpt.Name {
			return &regOpt, nil
		}
		if regOpt.Selector == RegOptSelectorImage && strings.HasPrefix(image.Name(), regOpt.Name) {
			return &regOpt, nil
		}
	}
	if len(name) == 0 {
		return nil, nil
	}
	return nil, errors.Errorf("%s not found", name)
}
