package signalrest

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
)

// Client represents an active signalrest notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifSignalRest
	meta model.Meta
}

// New creates a new signalrest notification instance
func New(config *model.NotifSignalRest, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "signalrest"
}

// Send creates and sends a signalrest notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	message, err := msg.New(msg.Options{
		Meta:         c.meta,
		Entry:        entry,
		TemplateBody: c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	_, bodyrender, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	body, err := json.Marshal(struct {
		Message    string   `json:"message"`
		Number     string   `json:"number"`
		Recipients []string `json:"recipients"`
	}{
		Message:    string(bodyrender),
		Number:     c.cfg.Number,
		Recipients: c.cfg.Recipients,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if len(c.cfg.Headers) > 0 {
		for key, value := range c.cfg.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	_, err = hc.Do(req)
	return err
}
