package teams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/internal/notif/notifier"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg       *model.NotifTeams
	userAgent string
}

// New creates a new webhook notification instance
func New(config *model.NotifTeams, userAgent string) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:       config,
			userAgent: userAgent,
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

	var textBuf bytes.Buffer
	textTpl := template.Must(template.New("text").Parse("Docker tag `{{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }}` {{ if (eq .Status \"new\") }}newly added{{ else }}updated{{ end }}."))
	if err := textTpl.Execute(&textBuf, entry); err != nil {
		return err
	}

	themeColor := "68CA00"
	if entry.Status == model.ImageStatusUpdate {
		themeColor = "0076D7"
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
		ThemeColor: themeColor,
		Summary:    textBuf.String(),
		Sections: []Sections{{
			ActivityTitle:    textBuf.String(),
			ActivitySubtitle: "Provider: " + entry.Provider,
			Facts: []Fact{
				{"Created", entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC")},
				{"Digest", entry.Manifest.Digest.String()},
				{"Platform", entry.Manifest.Platform},
			}}},
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
