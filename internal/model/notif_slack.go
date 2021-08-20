package model

// NotifSlackDefaultTemplateBody ...
const NotifSlackDefaultTemplateBody = "<!channel> Docker tag `{{ .Entry.Image }}` {{ if (eq .Entry.Status \"new\") }}available{{ else }}updated{{ end }}."

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	WebhookURL   string `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
	TemplateBody string `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifSlack) GetDefaults() *NotifSlack {
	n := &NotifSlack{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifSlack) SetDefaults() {
	s.TemplateBody = NotifSlackDefaultTemplateBody
}
