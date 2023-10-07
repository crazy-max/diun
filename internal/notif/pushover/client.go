package pushover

import (
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/gregdel/pushover"
	"github.com/pkg/errors"
)

// Client represents an active Pushover notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifPushover
	meta model.Meta
}

// New creates a new Pushover notification instance
func New(config *model.NotifPushover, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "pushover"
}

// Send creates and sends a Pushover notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Pushover notifier")
	}

	recipient, err := utl.GetSecret(c.cfg.Recipient, c.cfg.RecipientFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve recipient secret for Pushover notifier")
	}

	message, err := msg.New(msg.Options{
		Meta:          c.meta,
		Entry:         entry,
		TemplateTitle: c.cfg.TemplateTitle,
		TemplateBody:  c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	title, body, err := message.RenderHTML()
	if err != nil {
		return err
	}

	messageURL, err := msg.New(msg.Options{
		Meta:          c.meta,
		Entry:         entry,
		TemplateTitle: c.cfg.TemplateURLTitle,
		TemplateBody:  c.cfg.TemplateURL,
	})
	if err != nil {
		return err
	}

	urlTitle, url, err := messageURL.RenderMarkdown()
	if err != nil {
		return err
	}

	_, err = pushover.New(token).SendMessage(&pushover.Message{
		Title:     string(title),
		Message:   string(body),
		Priority:  c.cfg.Priority,
		Sound:     c.cfg.Sound,
		URLTitle:  string(urlTitle),
		URL:       string(url),
		Timestamp: time.Now().Unix(),
		HTML:      true,
	}, pushover.NewRecipient(recipient))

	return err
}
