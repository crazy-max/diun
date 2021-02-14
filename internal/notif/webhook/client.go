package webhook

import (
	"bytes"
	"net/http"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifWebhook
	meta model.Meta
}

// New creates a new webhook notification instance
func New(config *model.NotifWebhook, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "webhook"
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	body, err := message.RenderJSON()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(c.cfg.Method, c.cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if len(c.cfg.Headers) > 0 {
		for key, value := range c.cfg.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("User-Agent", c.meta.UserAgent)

	_, err = hc.Do(req)
	return err
}
