package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
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

	webhookURL, err := utl.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve webhook URL for Discord notifier")
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

	if len(c.cfg.Mentions) > 0 {
		for _, mention := range c.cfg.Mentions {
			content.WriteString(fmt.Sprintf("%s ", mention))
		}
	}
	content.WriteString(string(body))

	var fields []EmbedField
	if *c.cfg.RenderFields {
		fields = []EmbedField{
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
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(Message{
		Content:   content.String(),
		Username:  c.meta.Name,
		AvatarURL: c.meta.Logo,
		Embeds: []Embed{
			{
				Author: EmbedAuthor{
					Name:    c.meta.Name,
					URL:     c.meta.URL,
					IconURL: c.meta.Logo,
				},
				Fields: fields,
				Footer: EmbedFooter{
					Text: fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
				},
			},
		},
	}); err != nil {
		return err
	}

	u, err := url.Parse(webhookURL)
	if err != nil {
		return err
	}

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	hc := http.Client{}
	req, err := http.NewRequestWithContext(timeoutCtx, "POST", u.String(), dataBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, resp.Body)
	}

	return nil
}
