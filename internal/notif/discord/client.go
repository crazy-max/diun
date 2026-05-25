package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/internal/secret"
	"github.com/pkg/errors"
)

const discordMaxRateLimitAttempts = 3

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

	webhookURL, err := secret.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
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
			content.WriteString(mention)
			content.WriteString(" ")
		}
	}
	content.WriteString(string(body))

	var embeds []Embed
	if *c.cfg.RenderEmbeds {
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
		embeds = []Embed{
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
		}
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(Message{
		Content:   content.String(),
		Username:  c.meta.Name,
		AvatarURL: c.meta.Logo,
		Embeds:    embeds,
	}); err != nil {
		return err
	}
	payload := dataBuf.Bytes()

	u, err := url.Parse(webhookURL)
	if err != nil {
		return err
	}

	hc := http.Client{}
	for attempt := 1; attempt <= discordMaxRateLimitAttempts; attempt++ {
		cancelCtx, cancel := context.WithCancelCause(context.Background())
		timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent

		req, err := http.NewRequestWithContext(timeoutCtx, "POST", u.String(), bytes.NewReader(payload))
		if err != nil {
			cancel(errors.WithStack(context.Canceled))
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.meta.UserAgent)

		resp, err := hc.Do(req)
		if err != nil {
			cancel(errors.WithStack(context.Canceled))
			return err
		}

		dt, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); err == nil {
			err = closeErr
		}
		cancel(errors.WithStack(context.Canceled))
		if err != nil {
			return errors.Wrap(err, "cannot read Discord response")
		}

		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		err = errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(dt))
		if resp.StatusCode != http.StatusTooManyRequests || attempt == discordMaxRateLimitAttempts {
			return err
		}

		time.Sleep(discordRetryAfter(resp, dt))
	}

	return nil
}

// https://docs.discord.com/developers/topics/rate-limits
func discordRetryAfter(resp *http.Response, body []byte) time.Duration {
	if value := resp.Header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 {
			return time.Duration(seconds * float64(time.Second))
		}
		if retryAt, err := http.ParseTime(value); err == nil {
			delay := time.Until(retryAt)
			if delay > 0 {
				return delay
			}
		}
	}

	var errBody struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &errBody); err == nil && errBody.RetryAfter > 0 {
		return time.Duration(errBody.RetryAfter * float64(time.Second))
	}

	if value := resp.Header.Get("X-RateLimit-Reset-After"); value != "" {
		if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 {
			return time.Duration(seconds * float64(time.Second))
		}
	}

	return 5 * time.Second
}
