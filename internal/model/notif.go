package model

import (
	"github.com/crazy-max/diun/v4/pkg/registry"
)

// Defaults used for notification template
const (
	NotifDefaultTemplateTitle = `{{ .Entry.Image }} {{ if (eq .Entry.Status "new") }}is available{{ else }}has been updated{{ end }}`
	NotifDefaultTemplateBody  = `Docker tag {{ if .Entry.Image.HubLink }}[**{{ .Entry.Image }}**]({{ .Entry.Image.HubLink }}){{ else }}**{{ .Entry.Image }}**{{ end }} which you subscribed to through {{ .Entry.Provider }} provider {{ if (eq .Entry.Status "new") }}is available{{ else }}has been updated{{ end }} on {{ .Entry.Image.Domain }} registry (triggered by {{ .Meta.Hostname }} host).`
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
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Notif holds data necessary for notification configuration
type Notif struct {
	Amqp       *NotifAmqp       `yaml:"amqp,omitempty" json:"amqp,omitempty"`
	Discord    *NotifDiscord    `yaml:"discord,omitempty" json:"discord,omitempty"`
	Gotify     *NotifGotify     `yaml:"gotify,omitempty" json:"gotify,omitempty"`
	Mail       *NotifMail       `yaml:"mail,omitempty" json:"mail,omitempty"`
	Matrix     *NotifMatrix     `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	Mqtt       *NotifMqtt       `yaml:"mqtt,omitempty" json:"mqtt,omitempty"`
	Ntfy       *NotifNtfy       `yaml:"ntfy,omitempty" json:"ntfy,omitempty"`
	Pushover   *NotifPushover   `yaml:"pushover,omitempty" json:"pushover,omitempty"`
	RocketChat *NotifRocketChat `yaml:"rocketchat,omitempty" json:"rocketchat,omitempty"`
	Script     *NotifScript     `yaml:"script,omitempty" json:"script,omitempty"`
	SignalRest *NotifSignalRest `yaml:"signalrest,omitempty" json:"signalrest,omitempty"`
	Slack      *NotifSlack      `yaml:"slack,omitempty" json:"slack,omitempty"`
	Teams      *NotifTeams      `yaml:"teams,omitempty" json:"teams,omitempty"`
	Telegram   *NotifTelegram   `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	Webhook    *NotifWebhook    `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	HomeAssistant  *NotifHomeAssistant  `yaml:"homeassistant,omitempty" json:"homeassistant,omitempty"`
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
