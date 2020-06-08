package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/opencontainers/go-digest"
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

	body, err := json.Marshal(struct {
		Version  string        `json:"diun_version"`
		Status   string        `json:"status"`
		Provider string        `json:"provider"`
		Image    string        `json:"image"`
		HubLink  string        `json:"hub_link"`
		MIMEType string        `json:"mime_type"`
		Digest   digest.Digest `json:"digest"`
		Created  *time.Time    `json:"created"`
		Platform string        `json:"platform"`
	}{
		Version:  c.meta.Version,
		Status:   string(entry.Status),
		Provider: entry.Provider,
		Image:    entry.Image.String(),
		HubLink:  entry.Image.HubLink,
		MIMEType: entry.Manifest.MIMEType,
		Digest:   entry.Manifest.Digest,
		Created:  entry.Manifest.Created,
		Platform: entry.Manifest.Platform,
	})
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
