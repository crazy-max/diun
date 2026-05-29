package webhook

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/crazy-max/diun/v4/internal/httputil"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/pkg/errors"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifWebhook
	meta model.Meta
}

// New creates a new webhook notification instance
func New(config *model.NotifWebhook, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "webhook"
}

// Send creates and sends a webhook notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	body, err := message.RenderJSON()
	if err != nil {
		return err
	}

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Webhook notifier")
	}

	req, err := http.NewRequestWithContext(timeoutCtx, c.cfg.Method, c.cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if len(c.cfg.Headers) > 0 {
		for key, value := range c.cfg.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("User-Agent", c.meta.UserAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "cannot read Webhook error response")
		}
		return errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
