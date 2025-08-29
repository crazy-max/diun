package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

type NotifElasticsearch struct {
	Address            string         `yaml:"address,omitempty" json:"address,omitempty" validate:"required"`
	Username           string         `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile       string         `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password           string         `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile       string         `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	Client             string         `yaml:"client,omitempty" json:"client,omitempty" validate:"required"`
	Index              string         `yaml:"index,omitempty" json:"index,omitempty" validate:"required"`
	Timeout            *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	InsecureSkipVerify bool           `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifElasticsearch) GetDefaults() *NotifElasticsearch {
	n := &NotifElasticsearch{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifElasticsearch) SetDefaults() {
	s.Address = "http://localhost:9200"
	s.Client = "diun"
	s.Index = "diun-notifications"
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.InsecureSkipVerify = false
}
