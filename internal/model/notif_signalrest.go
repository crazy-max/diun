package model

import (
	"time"

	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifSignalRestDefaultTemplateBody ...
const NotifSignalRestDefaultTemplateBody = `Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider {{ if (eq .Entry.Status "new") }}is available{{ else }}has been updated{{ end }} on {{ .Entry.Image.Domain }} registry (triggered by {{ .Meta.Hostname }} host).`

// NotifSignalRest holds SignalRest notification configuration details
type NotifSignalRest struct {
	Endpoint     string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Number       string            `yaml:"number,omitempty" json:"method,omitempty" validate:"required"`
	Recipients   []string          `yaml:"recipients,omitempty" json:"recipients,omitempty" validate:"omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" validate:"omitempty"`
	Timeout      *time.Duration    `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required"`
	TemplateBody string            `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifSignalRest) GetDefaults() *NotifSignalRest {
	n := &NotifSignalRest{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifSignalRest) SetDefaults() {
	s.Timeout = utl.NewDuration(10 * time.Second)
	s.Endpoint = "http://localhost:8080/v2/send"
	s.TemplateBody = NotifSignalRestDefaultTemplateBody
}
