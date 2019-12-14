package model

import (
	"github.com/crazy-max/diun/pkg/docker"
	"github.com/crazy-max/diun/pkg/docker/registry"
)

// Notif holds data necessary for notification configuration
type Notif struct {
	Mail    Mail    `yaml:"mail,omitempty"`
	Webhook Webhook `yaml:"webhook,omitempty"`
}

// NotifEntry represents a notification entry
type NotifEntry struct {
	Status   ImageStatus     `json:"status,omitempty"`
	Provider string          `json:"provider,omitempty"`
	Image    registry.Image  `json:"image,omitempty"`
	Manifest docker.Manifest `json:"manifest,omitempty"`
}
