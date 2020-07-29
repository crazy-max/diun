package notif

import (
	"strings"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/amqp"
	"github.com/crazy-max/diun/v4/internal/notif/discord"
	"github.com/crazy-max/diun/v4/internal/notif/gotify"
	"github.com/crazy-max/diun/v4/internal/notif/mail"
	"github.com/crazy-max/diun/v4/internal/notif/matrix"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/internal/notif/rocketchat"
	"github.com/crazy-max/diun/v4/internal/notif/script"
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
	if config.Discord != nil {
		c.notifiers = append(c.notifiers, discord.New(config.Discord, meta))
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
	if config.RocketChat != nil {
		c.notifiers = append(c.notifiers, rocketchat.New(config.RocketChat, meta))
	}
	if config.Script != nil {
		c.notifiers = append(c.notifiers, script.New(config.Script, meta))
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
			log.Error().Err(err).Str("image", entry.Image.String()).Msgf("%s notification failed", strings.Title(n.Name()))
		}
	}
}
