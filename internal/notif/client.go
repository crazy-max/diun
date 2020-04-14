package notif

import (
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/amqp"
	"github.com/crazy-max/diun/internal/notif/gotify"
	"github.com/crazy-max/diun/internal/notif/mail"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/crazy-max/diun/internal/notif/rocketchat"
	"github.com/crazy-max/diun/internal/notif/slack"
	"github.com/crazy-max/diun/internal/notif/telegram"
	"github.com/crazy-max/diun/internal/notif/webhook"
	"github.com/rs/zerolog/log"
)

// Client represents an active webhook notification object
type Client struct {
	cfg       model.Notif
	app       model.App
	notifiers []notifier.Notifier
}

// New creates a new notification instance
func New(config model.Notif, app model.App, userAgent string) (*Client, error) {
	var c = &Client{
		cfg:       config,
		app:       app,
		notifiers: []notifier.Notifier{},
	}

	// Add notifiers
	if config.Gotify.Enable {
		c.notifiers = append(c.notifiers, gotify.New(config.Gotify, app, userAgent))
	}
	if config.Mail.Enable {
		c.notifiers = append(c.notifiers, mail.New(config.Mail, app))
	}
	if config.RocketChat.Enable {
		c.notifiers = append(c.notifiers, rocketchat.New(config.RocketChat, app, userAgent))
	}
	if config.Slack.Enable {
		c.notifiers = append(c.notifiers, slack.New(config.Slack, app))
	}
	if config.Telegram.Enable {
		c.notifiers = append(c.notifiers, telegram.New(config.Telegram, app))
	}
	if config.Webhook.Enable {
		c.notifiers = append(c.notifiers, webhook.New(config.Webhook, app, userAgent))
	}
	if config.Amqp.Enable {
		c.notifiers = append(c.notifiers, amqp.New(config.Amqp, app))
	}

	log.Debug().Msgf("%d notifier(s) created", len(c.notifiers))
	return c, nil
}

// Send creates and sends notifications to notifiers
func (c *Client) Send(entry model.NotifEntry) {
	for _, n := range c.notifiers {
		log.Debug().Str("image", entry.Image.String()).Msgf("Sending %s notification...", n.Name())
		if err := n.Send(entry); err != nil {
			log.Error().Err(err).Str("image", entry.Image.String()).Msgf("%s notification failed", n.Name())
		}
	}
}
