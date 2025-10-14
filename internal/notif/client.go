package notif

import (
	"regexp"
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/amqp"
	"github.com/crazy-max/diun/v4/internal/notif/apprise"
	"github.com/crazy-max/diun/v4/internal/notif/discord"
	"github.com/crazy-max/diun/v4/internal/notif/elasticsearch"
	"github.com/crazy-max/diun/v4/internal/notif/gotify"
	"github.com/crazy-max/diun/v4/internal/notif/mail"
	"github.com/crazy-max/diun/v4/internal/notif/matrix"
	"github.com/crazy-max/diun/v4/internal/notif/mqtt"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/internal/notif/ntfy"
	"github.com/crazy-max/diun/v4/internal/notif/pushover"
	"github.com/crazy-max/diun/v4/internal/notif/rocketchat"
	"github.com/crazy-max/diun/v4/internal/notif/script"
	"github.com/crazy-max/diun/v4/internal/notif/signalrest"
	"github.com/crazy-max/diun/v4/internal/notif/slack"
	"github.com/crazy-max/diun/v4/internal/notif/teams"
	"github.com/crazy-max/diun/v4/internal/notif/telegram"
	"github.com/crazy-max/diun/v4/internal/notif/webhook"
	"github.com/rs/zerolog/log"
)

// Client represents an active webhook notification object
type Client struct {
	cfg       *model.Notif
	meta      model.Meta
	notifiers []notifier.Notifier
}

// New creates a new notification instance
func New(config *model.Notif, meta model.Meta) (*Client, error) {
	var c = &Client{
		cfg:       config,
		meta:      meta,
		notifiers: []notifier.Notifier{},
	}

	if config == nil {
		log.Warn().Msg("No notifier available")
		return c, nil
	}

	// Add notifiers
	if config.Amqp != nil {
		c.notifiers = append(c.notifiers, amqp.New(config.Amqp, meta))
	}
	if config.Apprise != nil {
		c.notifiers = append(c.notifiers, apprise.New(config.Apprise, meta))
	}
	if config.Discord != nil {
		c.notifiers = append(c.notifiers, discord.New(config.Discord, meta))
	}
	if config.Elasticsearch != nil {
		c.notifiers = append(c.notifiers, elasticsearch.New(config.Elasticsearch, meta))
	}
	if config.Gotify != nil {
		c.notifiers = append(c.notifiers, gotify.New(config.Gotify, meta))
	}
	if config.Mail != nil {
		c.notifiers = append(c.notifiers, mail.New(config.Mail, meta))
	}
	if config.Matrix != nil {
		c.notifiers = append(c.notifiers, matrix.New(config.Matrix, meta))
	}
	if config.Mqtt != nil {
		c.notifiers = append(c.notifiers, mqtt.New(config.Mqtt, meta))
	}
	if config.Ntfy != nil {
		c.notifiers = append(c.notifiers, ntfy.New(config.Ntfy, meta))
	}
	if config.Pushover != nil {
		c.notifiers = append(c.notifiers, pushover.New(config.Pushover, meta))
	}
	if config.RocketChat != nil {
		c.notifiers = append(c.notifiers, rocketchat.New(config.RocketChat, meta))
	}
	if config.Script != nil {
		c.notifiers = append(c.notifiers, script.New(config.Script, meta))
	}
	if config.SignalRest != nil {
		c.notifiers = append(c.notifiers, signalrest.New(config.SignalRest, meta))
	}
	if config.Slack != nil {
		c.notifiers = append(c.notifiers, slack.New(config.Slack, meta))
	}
	if config.Teams != nil {
		c.notifiers = append(c.notifiers, teams.New(config.Teams, meta))
	}
	if config.Telegram != nil {
		c.notifiers = append(c.notifiers, telegram.New(config.Telegram, meta))
	}
	if config.Webhook != nil {
		c.notifiers = append(c.notifiers, webhook.New(config.Webhook, meta))
	}

	log.Debug().Msgf("%d notifier(s) created", len(c.notifiers))
	return c, nil
}

// Send creates and sends notifications to notifiers
func (c *Client) Send(entry model.NotifEntry) {
	for _, n := range c.notifiers {
		log.Debug().Str("image", entry.Image.String()).Msgf("Sending %s notification...", n.Name())
		if err := n.Send(entry); err != nil {
			log.Error().Str("image", entry.Image.String()).Msgf("%s notification failed: %s", strings.Title(n.Name()), SanitizeUrlTokens(err)) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support
		}
	}
}

// SanitizeUrlTokens redacts auth tokens in URLs from error messages
func SanitizeUrlTokens(err error) string {
	if err == nil {
		return ""
	}
	params := []string{"token", "apikey", "api_key", "access_token", "auth", "authorization", "jwt", "sessionid", "session_id", "password", "secret", "key", "code"}
	pattern := `([?&](` + strings.Join(params, "|") + `)=)[^&"\s]+` // scan ? or & followed by one of the param names and =, then redact until &, whitespace, or " (end of URL)
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(err.Error(), `$1[REDACTED]`) // leave param name, redact secret value
}

// List returns created notifiers
func (c *Client) List() []notifier.Notifier {
	return c.notifiers
}
