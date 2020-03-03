package teams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg       model.NotifTeams
	app       model.App
	userAgent string
}

// New creates a new webhook notification instance
func New(config model.NotifTeams, app model.App, userAgent string) notifier.Notifier {
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
	return "teams"
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: time.Duration(10) * time.Second,
	}

	body, err := json.Marshal(struct {
		Type       string `json:"@type"`
		Text       string `json:"text"`
		ThemeColor string `json:"themeColor"`
		Summary    string `json:"summary"`
	}{
		Type:       "MessageCard",
		Text:       entry.Image.String(),
		ThemeColor: "0076D7",
		Summary:    string(entry.Status),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.cfg.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	req.Header.Set("User-Agent", c.userAgent)

	_, err = hc.Do(req)
	return err
}
