package model

// NotifPushover holds Pushover notification configuration details
type NotifPushover struct {
	Token             string `yaml:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	TokenFile         string `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Recipient         string `yaml:"recipient,omitempty" json:"recipient,omitempty" validate:"omitempty"`
	RecipientFile     string `yaml:"recipientFile,omitempty" json:"recipientFile,omitempty" validate:"omitempty,file"`
	Priority          int    `yaml:"priority,omitempty" json:"priority,omitempty" validate:"omitempty,min=-2,max=2"`
	Sound             string `yaml:"sound,omitempty" json:"sound,omitempty" validate:"omitempty"`
	TemplateTitle     string `yaml:"templateTitle,omitempty" json:"templateTitle,omitempty" validate:"required"`
	TemplateBody      string `yaml:"templateBody,omitempty" json:"templateBody,omitempty" validate:"required"`
	TemplateURL       string `yaml:"templateUrl,omitempty" json:"templateUrl,omitempty" validate:"required"`
	TemplateURLTitle  string `yaml:"templateUrlTitle,omitempty" json:"templateUrlTitle,omitempty" validate:"required"`
}

// GetDefaults gets the default values
func (s *NotifPushover) GetDefaults() *NotifPushover {
	n := &NotifPushover{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifPushover) SetDefaults() {
	s.TemplateTitle = NotifDefaultTemplateTitle
	s.TemplateBody = NotifDefaultTemplateBody
	s.TemplateURLTitle = NotifDefaultTemplateURLTitle
	s.TemplateURL = NotifDefaultTemplateURL
}
