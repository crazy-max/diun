package model

import (
	"github.com/crazy-max/diun/v4/pkg/utl"
)

// NotifMailDefaultTemplateBody ...
const NotifMailDefaultTemplateBody = `Docker tag {{ if .Entry.Image.HubLink }}[**{{ .Entry.Image }}**]({{ .Entry.Image.HubLink }}){{ else }}**{{ .Entry.Image }}**{{ end }}
which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}
on {{ .Meta.Hostname }}.

This image has been {{ if (eq .Entry.Status "new") }}created{{ else }}updated{{ end }} at
<code>{{ .Entry.Manifest.Created.Format "Jan 02, 2006 15:04:05 UTC" }}</code> with digest <code>{{ .Entry.Manifest.Digest }}</code>
for <code>{{ .Entry.Manifest.Platform }}</code> platform.

Need help, or have questions? Go to {{ .Meta.URL }} and leave an issue.`

// NotifMail holds mail notification configuration details
type NotifMail struct {
	Host               string   `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port               int      `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
	SSL                *bool    `yaml:"ssl,omitempty" json:"ssl,omitempty" validate:"required"`
	InsecureSkipVerify *bool    `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty" validate:"required"`
	LocalName          string   `yaml:"localName,omitempty" json:"localName,omitempty" validate:"omitempty"`
	Username           string   `yaml:"username,omitempty" json:"username,omitempty" validate:"omitempty"`
	UsernameFile       string   `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password           string   `yaml:"password,omitempty" json:"password,omitempty" validate:"omitempty"`
	PasswordFile       string   `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	From               string   `yaml:"from,omitempty" json:"from,omitempty" validate:"required,email"`
	To                 []string `yaml:"to,omitempty" json:"to,omitempty" validate:"required"`
	TemplateTitle      string   `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody       string   `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifMail) GetDefaults() *NotifMail {
	n := &NotifMail{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifMail) SetDefaults() {
	s.Host = "localhost"
	s.Port = 25
	s.SSL = utl.NewFalse()
	s.InsecureSkipVerify = utl.NewFalse()
	s.LocalName = "localhost"
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifMailDefaultTemplateBody
}
