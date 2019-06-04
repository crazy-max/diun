package model

import "github.com/crazy-max/diun/pkg/registry"

// Notif holds data necessary for notification configuration
type Notif struct {
	Mail    Mail    `yaml:"mail,omitempty"`
	Webhook Webhook `yaml:"webhook,omitempty"`
}

// NotifEntry represents a notification entry
type NotifEntry struct {
	Status   ImageStatus      `json:"status,omitempty"`
	Image    registry.Image   `json:"image,omitempty"`
	Analysis registry.Inspect `json:"analysis,omitempty"`
}
