package model

import (
	"github.com/crazy-max/diun/v4/pkg/registry"
)

// NotifEntries represents a list of notification entries
type NotifEntries struct {
	Entries       []NotifEntry
	CountNew      int
	CountUpdate   int
	CountUnchange int
	CountSkip     int
	CountError    int
	CountTotal    int
}

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
	Discord    *NotifDiscord    `yaml:"discord,omitempty" json:"discord,omitempty"`
	Gotify     *NotifGotify     `yaml:"gotify,omitempty" json:"gotify,omitempty"`
	Mail       *NotifMail       `yaml:"mail,omitempty" json:"mail,omitempty"`
	Matrix     *NotifMatrix     `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	Mqtt       *NotifMqtt       `yaml:"mqtt,omitempty" json:"mqtt,omitempty"`
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

// Add adds a new notif entry
func (s *NotifEntries) Add(entry NotifEntry) {
	s.Entries = append(s.Entries, entry)
	switch entry.Status {
	case ImageStatusNew:
		s.CountNew++
		s.CountTotal++
	case ImageStatusUpdate:
		s.CountUpdate++
		s.CountTotal++
	case ImageStatusUnchange:
		s.CountUnchange++
		s.CountTotal++
	case ImageStatusSkip:
		s.CountSkip++
		s.CountTotal++
	case ImageStatusError:
		s.CountError++
		s.CountTotal++
	}
}
