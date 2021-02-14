package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
)

// Client represents an active discord notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifDiscord
	meta model.Meta
}

// New creates a new discord notification instance
func New(config *model.NotifDiscord, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "discord"
}

// Send creates and sends a discord notification with an entry
// https://discord.com/developers/docs/resources/webhook#execute-webhook
func (c *Client) Send(entry model.NotifEntry) error {
	var content bytes.Buffer

	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	title, text, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	if len(c.cfg.Mentions) > 0 {
		for _, mention := range c.cfg.Mentions {
			content.WriteString(fmt.Sprintf("%s ", mention))
		}
	}
	content.WriteString(title)

	fields := []EmbedField{
		{
			Name:  "Hostname",
			Value: c.meta.Hostname,
		},
		{
			Name:  "Provider",
			Value: entry.Provider,
		},
		{
			Name:  "Created",
			Value: entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC"),
		},
		{
			Name:  "Digest",
			Value: entry.Manifest.Digest.String(),
		},
		{
			Name:  "Platform",
			Value: entry.Manifest.Platform,
		},
	}
	if len(entry.Image.HubLink) > 0 {
		fields = append(fields, EmbedField{
			Name:  "HubLink",
			Value: entry.Image.HubLink,
		})
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(Message{
		Content:   content.String(),
		Username:  c.meta.Name,
		AvatarURL: c.meta.Logo,
		Embeds: []Embed{
			{
				Description: string(text),
				Footer: EmbedFooter{
					Text:    fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
					IconURL: c.meta.Logo,
				},
				Author: EmbedAuthor{
					Name:    c.meta.Name,
					URL:     c.meta.URL,
					IconURL: c.meta.Logo,
				},
				Fields: fields,
			},
		},
	}); err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.WebhookURL)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), dataBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, resp.Body)
	}

	return nil
}
