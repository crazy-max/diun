package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/opencontainers/go-digest"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg       *model.NotifWebhook
	app       model.App
	userAgent string
}

// New creates a new webhook notification instance
func New(config *model.NotifWebhook, app model.App, userAgent string) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:       config,
			app:       app,
			userAgent: userAgent,
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
		Timeout: time.Duration(c.cfg.Timeout) * time.Second,
	}

	body, err := json.Marshal(struct {
		Version  string        `json:"diun_version"`
		Status   string        `json:"status"`
		Provider string        `json:"provider"`
		Image    string        `json:"image"`
		MIMEType string        `json:"mime_type"`
		Digest   digest.Digest `json:"digest"`
		Created  *time.Time    `json:"created"`
		Platform string        `json:"platform"`
	}{
		Version:  c.app.Version,
		Status:   string(entry.Status),
		Provider: entry.Provider,
		Image:    entry.Image.String(),
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

	req.Header.Set("User-Agent", c.userAgent)

	_, err = hc.Do(req)
	return err
}
