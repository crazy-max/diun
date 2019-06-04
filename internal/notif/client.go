package notif

import (
	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/mail"
	"github.com/crazy-max/diun/internal/notif/notifier"
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
func New(config model.Notif, app model.App) (*Client, error) {
	var c = &Client{
		cfg:       config,
		app:       app,
		notifiers: []notifier.Notifier{},
	}

	// Add notifiers
	if config.Mail.Enable {
		c.notifiers = append(c.notifiers, mail.New(config.Mail, app))
	}
	if config.Webhook.Enable {
		c.notifiers = append(c.notifiers, webhook.New(config.Webhook, app))
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
