package slack

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

// Client represents an active slack notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifSlack
	meta model.Meta
}

// New creates a new slack notification instance
func New(config *model.NotifSlack, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "slack"
}

// Send creates and sends a slack notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	webhookURL, err := utl.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve webhook URL for Slack notifier")
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

	var fields []slack.AttachmentField
	if *c.cfg.RenderFields {
		fields = []slack.AttachmentField{
			{
				Title: "Hostname",
				Value: c.meta.Hostname,
				Short: false,
			},
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
				Value: entry.Manifest.Platform,
				Short: false,
			},
		}
		if len(entry.Image.HubLink) > 0 {
			fields = append(fields, slack.AttachmentField{
				Title: "HubLink",
				Value: entry.Image.HubLink,
				Short: false,
			})
		}
	}

	color := "#4caf50"
	if entry.Status == model.ImageStatusUpdate {
		color = "#0054ca"
	}

	return slack.PostWebhook(webhookURL, &slack.WebhookMessage{
		Attachments: []slack.Attachment{
			{
				Color:         color,
				AuthorName:    c.meta.Name,
				AuthorSubname: "github.com/crazy-max/diun",
				AuthorLink:    c.meta.URL,
				AuthorIcon:    c.meta.Logo,
				Text:          string(body),
				Footer:        fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
				Fields:        fields,
				Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
			},
		},
	})
}
