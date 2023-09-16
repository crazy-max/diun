package ntfy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Client represents an active ntfy notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifNtfy
	meta model.Meta
}

// New creates a new ntfy notification instance
func New(config *model.NotifNtfy, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "ntfy"
}

// Send creates and sends a ntfy notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	message, err := msg.New(msg.Options{
		Meta:          c.meta,
		Entry:         entry,
		TemplateTitle: c.cfg.TemplateTitle,
		TemplateBody:  c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	title, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(struct {
		Topic    string   `json:"topic"`
		Message  string   `json:"message"`
		Title    string   `json:"title"`
		Priority int      `json:"priority"`
		Tags     []string `json:"tags"`
	}{
		Topic:    c.cfg.Topic,
		Message:  string(body),
		Title:    string(title),
		Priority: c.cfg.Priority,
		Tags:     c.cfg.Tags,
	}); err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}

	q := u.Query()
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), dataBuf)
	if err != nil {
		return err
	}

	if token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile); err == nil && token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error            string `json:"error"`
			ErrorCode        int    `json:"errorCode"`
			ErrorDescription string `json:"errorDescription"`
		}
		err := json.NewDecoder(resp.Body).Decode(&errBody)
		if err != nil {
			return err
		}
		return errors.Errorf("%d %s: %s", errBody.ErrorCode, errBody.Error, errBody.ErrorDescription)
	}

	return nil
}
