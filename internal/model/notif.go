package model

import (
	"github.com/crazy-max/diun/v3/pkg/registry"
)

// NotifEntry represents a notification entry
type NotifEntry struct {
	Status   ImageStatus       `json:"status,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Image    registry.Image    `json:"image,omitempty"`
	Manifest registry.Manifest `json:"manifest,omitempty"`
}

// Notif holds data necessary for notification configuration
type Notif struct {
	Amqp       *NotifAmqp       `yaml:"amqp,omitempty" json:"amqp,omitempty"`
	Gotify     *NotifGotify     `yaml:"gotify,omitempty" json:"gotify,omitempty"`
	Mail       *NotifMail       `yaml:"mail,omitempty" json:"mail,omitempty"`
	RocketChat *NotifRocketChat `yaml:"rocketchat,omitempty" json:"rocketchat,omitempty"`
	Script     *NotifScript     `yaml:"script,omitempty" json:"script,omitempty"`
	Slack      *NotifSlack      `yaml:"slack,omitempty" json:"slack,omitempty"`
	Teams      *NotifTeams      `yaml:"teams,omitempty" json:"teams,omitempty"`
	Telegram   *NotifTelegram   `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	Webhook    *NotifWebhook    `yaml:"webhook,omitempty" json:"webhook,omitempty"`
}

// GetDefaults gets the default values
func (s *Notif) GetDefaults() *Notif {
	return nil
}

// SetDefaults sets the default values
func (s *Notif) SetDefaults() {
	// noop
}
