package model

import (
	"net/mail"
	"os/exec"
	"strings"
	"time"

	"github.com/crazy-max/diun/v3/pkg/registry"
	"github.com/crazy-max/diun/v3/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
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
	Amqp       *NotifAmqp       `yaml:"amqp,omitempty"`
	Gotify     *NotifGotify     `yaml:"gotify,omitempty"`
	Mail       *NotifMail       `yaml:"mail,omitempty"`
	RocketChat *NotifRocketChat `yaml:"rocketchat,omitempty"`
	Script     *NotifScript     `yaml:"script,omitempty"`
	Slack      *NotifSlack      `yaml:"slack,omitempty"`
	Teams      *NotifTeams      `yaml:"teams,omitempty"`
	Telegram   *NotifTelegram   `yaml:"telegram,omitempty"`
	Webhook    *NotifWebhook    `yaml:"webhook,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *Notif) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Notif
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

// NotifAmqp holds amqp notification configuration details
type NotifAmqp struct {
	Username     string `yaml:"username,omitempty"`
	UsernameFile string `yaml:"username_file,omitempty"`
	Password     string `yaml:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty"`
	Host         string `yaml:"host,omitempty"`
	Port         int    `yaml:"port,omitempty"`
	Queue        string `yaml:"queue,omitempty"`
	Exchange     string `yaml:"exchange,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifAmqp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifAmqp
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, NotifAmqp{
		Host: "localhost",
		Port: 5672,
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for amqp notif")
	}

	return nil
}

// NotifGotify holds gotify notification configuration details
type NotifGotify struct {
	Endpoint string         `yaml:"endpoint,omitempty"`
	Token    string         `yaml:"token,omitempty"`
	Priority int            `yaml:"priority,omitempty"`
	Timeout  *time.Duration `yaml:"timeout,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifGotify) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifGotify
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, NotifGotify{
		Timeout: utl.NewDuration(10 * time.Second),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for gotify notif")
	}

	return nil
}

// NotifMail holds mail notification configuration details
type NotifMail struct {
	Host               string `yaml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	SSL                *bool  `yaml:"ssl,omitempty"`
	InsecureSkipVerify *bool  `yaml:"insecure_skip_verify,omitempty"`
	Username           string `yaml:"username,omitempty"`
	UsernameFile       string `yaml:"username_file,omitempty"`
	Password           string `yaml:"password,omitempty"`
	PasswordFile       string `yaml:"password_file,omitempty"`
	From               string `yaml:"from,omitempty"`
	To                 string `yaml:"to,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifMail) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifMail
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(s.From); err != nil {
		return errors.Wrap(err, "cannot parse sender mail address")
	}
	if _, err := mail.ParseAddress(s.To); err != nil {
		return errors.Wrap(err, "cannot parse recipient mail address")
	}

	if err := mergo.Merge(s, NotifMail{
		Host:               "localhost",
		Port:               25,
		SSL:                utl.NewFalse(),
		InsecureSkipVerify: utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for mail notif")
	}

	return nil
}

// NotifRocketChat holds Rocket.Chat notification configuration details
type NotifRocketChat struct {
	Endpoint string         `yaml:"endpoint,omitempty"`
	Channel  string         `yaml:"channel,omitempty"`
	UserID   string         `yaml:"user_id,omitempty"`
	Token    string         `yaml:"token,omitempty"`
	Timeout  *time.Duration `yaml:"timeout,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifRocketChat) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifRocketChat
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if err := mergo.Merge(s, NotifRocketChat{
		Timeout: utl.NewDuration(10 * time.Second),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for rocketchat notif")
	}

	return nil
}

// NotifScript holds script notification configuration details
type NotifScript struct {
	Cmd  string   `yaml:"cmd,omitempty"`
	Args []string `yaml:"args,omitempty"`
	Dir  string   `yaml:"dir,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifScript) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifScript
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if s.Cmd == "" {
		return errors.New("command required for script provider")
	}

	if _, err := exec.LookPath(s.Cmd); err != nil {
		return errors.Wrap(err, "command not found for script provider")
	}

	return nil
}

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifSlack) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifSlack
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

// NotifTeams holds Teams notification configuration details
type NotifTeams struct {
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifTeams) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifTeams
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

// NotifTelegram holds Telegram notification configuration details
type NotifTelegram struct {
	Token   string  `yaml:"token,omitempty"`
	ChatIDs []int64 `yaml:"chat_ids,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *NotifTelegram) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifTelegram
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}

// NotifWebhook holds webhook notification configuration details
type NotifWebhook struct {
	Endpoint string            `yaml:"endpoint,omitempty"`
	Method   string            `yaml:"method,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Timeout  *time.Duration    `yaml:"timeout,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (s *NotifWebhook) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NotifWebhook
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if len(s.Headers) == 0 {
		return nil
	}

	headers := make(map[string]string)
	for key, value := range s.Headers {
		headers[strings.ToLower(key)] = value
	}
	s.Headers = headers

	if err := mergo.Merge(s, NotifWebhook{
		Method:  "GET",
		Timeout: utl.NewDuration(10 * time.Second),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for webhook notif")
	}

	return nil
}
