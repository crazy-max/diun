package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifWebhook holds webhook notification configuration details
type NotifWebhook struct {
	Endpoint       string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Method         string            `yaml:"method,omitempty" json:"method,omitempty" validate:"required"`
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" validate:"omitempty"`
	Timeout        *time.Duration    `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TLSSkipVerify  bool              `yaml:"tlsSkipVerify,omitempty" json:"tlsSkipVerify,omitempty" validate:"omitempty"`
	TLSCACertFiles []string          `yaml:"tlsCaCertFiles,omitempty" json:"tlsCaCertFiles,omitempty" validate:"omitempty"`
}

// GetDefaults gets the default values
func (s *NotifWebhook) GetDefaults() *NotifWebhook {
	n := &NotifWebhook{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifWebhook) SetDefaults() {
	s.Method = "GET"
	s.Timeout = utl.NewDuration(10 * time.Second)
}
