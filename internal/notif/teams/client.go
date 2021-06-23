package teams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifTeams
	meta model.Meta
}

// New creates a new webhook notification instance
func New(config *model.NotifTeams, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "teams"
}

// Sections is grouping data together containing title, subtitle and facts and creating a nested json element
type Sections struct {
	ActivityTitle    string `json:"activityTitle"`
	ActivitySubtitle string `json:"activitySubtitle"`
	Facts            []Fact `json:"facts"`
}

// Fact is grouping data togheter to create a nested json element containg a name and an associated value
type Fact struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: time.Duration(10) * time.Second,
	}

	message, err := msg.New(msg.Options{
		Meta:         c.meta,
		Entry:        entry,
		TemplateBody: c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	_, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	themeColor := "68CA00"
	if entry.Status == model.ImageStatusUpdate {
		themeColor = "0076D7"
	}

	jsonBody, err := json.Marshal(struct {
		Type       string     `json:"@type"`
		Context    string     `json:"@context"`
		ThemeColor string     `json:"themeColor"`
		Summary    string     `json:"summary"`
		Sections   []Sections `json:"sections"`
	}{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		ThemeColor: themeColor,
		Summary:    string(body),
		Sections: []Sections{{
			ActivityTitle:    string(body),
			ActivitySubtitle: "Provider: " + entry.Provider,
			Facts: []Fact{
				{"Hostname", c.meta.Hostname},
				{"Provider", entry.Provider},
				{"Created", entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC")},
				{"Digest", entry.Manifest.Digest.String()},
				{"Platform", entry.Manifest.Platform},
			}}},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.cfg.WebhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	_, err = hc.Do(req)
	return err
}
