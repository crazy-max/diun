package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/nlopes/slack"
)

// Client represents an active slack notification object
type Client struct {
	*notifier.Notifier
	cfg model.NotifSlack
	app model.App
}

// New creates a new slack notification instance
func New(config model.NotifSlack, app model.App) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "slack"
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	var textBuf bytes.Buffer
	textTpl := template.Must(template.New("text").Parse("<!channel> Docker tag `{{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }}` {{ if (eq .Status \"new\") }}newly added{{ else }}updated{{ end }}."))
	if err := textTpl.Execute(&textBuf, entry); err != nil {
		return err
	}

	color := "#4caf50"
	if entry.Status == model.ImageStatusUpdate {
		color = "#0054ca"
	}

	return slack.PostWebhook(c.cfg.WebhookURL, &slack.WebhookMessage{
		Attachments: []slack.Attachment{slack.Attachment{
			Color:         color,
			AuthorName:    "Diun",
			AuthorSubname: "github.com/crazy-max/diun",
			AuthorLink:    "https://github.com/crazy-max/diun",
			AuthorIcon:    "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
			Text:          textBuf.String(),
			Footer:        fmt.Sprintf("%s © %d %s %s", c.app.Author, time.Now().Year(), c.app.Name, c.app.Version),
			Fields: []slack.AttachmentField{
				{
					Title: "Provider",
					Value: entry.Provider,
					Short: false,
				},
				{
					Title: "Created",
					Value: entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC"),
					Short: false,
				},
				{
					Title: "Digest",
					Value: entry.Manifest.Digest.String(),
					Short: false,
				},
				{
					Title: "Platform",
					Value: fmt.Sprintf("%s/%s", entry.Manifest.Os, entry.Manifest.Architecture),
					Short: false,
				},
			},
			Ts: json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		}},
	})
}
