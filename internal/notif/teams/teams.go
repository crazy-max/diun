package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Sections struct {
	ActivityTitle    string `json:"activityTitle"`
	ActivitySubtitle string `json:"activitySubtitle"`
	Facts            []Fact `json:"facts"`
}

type Fact struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: time.Duration(10) * time.Second,
	}

	var body, err = json.Marshal(struct {
		Type       string     `json:"@type"`
		Context    string     `json:"@context"`
		ThemeColor string     `json:"themeColor"`
		Summary    string     `json:"summary"`
		Sections   []Sections `json:"sections"`
	}{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "0076D7",
		Summary:    "Docker tag " + entry.Image.String() + " newly added",
		Sections: []Sections{{
			ActivityTitle:    "Docker tag " + entry.Image.String() + " newly added",
			ActivitySubtitle: "Provider: " + entry.Provider,
			Facts: []Fact{
				{"Created", entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC")},
				{"Digest", entry.Manifest.Digest.String()},
				{"Plattform", fmt.Sprintf("%s/%s", entry.Manifest.Os, entry.Manifest.Architecture)},
			},}},
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
