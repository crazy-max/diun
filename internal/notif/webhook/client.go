package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/opencontainers/go-digest"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg model.Webhook
	app model.App
}

// New creates a new webhook notification instance
func New(config model.Webhook, app model.App) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
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
		Version      string        `json:"diun_version,omitempty"`
		Status       string        `json:"status,omitempty"`
		Image        string        `json:"image,omitempty"`
		MIMEType     string        `json:"mime_type,omitempty"`
		Digest       digest.Digest `json:"digest,omitempty"`
		Date         *time.Time    `json:"date,omitempty"`
		Architecture string        `json:"architecture,omitempty"`
		Os           string        `json:"os,omitempty"`
	}{
		Version:      c.app.Version,
		Status:       string(entry.Status),
		Image:        entry.Image.String(),
		MIMEType:     entry.Analysis.MIMEType,
		Digest:       entry.Analysis.Digest,
		Date:         entry.Analysis.Created,
		Architecture: entry.Analysis.Architecture,
		Os:           entry.Analysis.Os,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(c.cfg.Method, c.cfg.Endpoint, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}

	if len(c.cfg.Headers) > 0 {
		for key, value := range c.cfg.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("User-Agent", fmt.Sprintf("%s %s", c.app.Name, c.app.Version))

	_, err = hc.Do(req)
	return err
}
