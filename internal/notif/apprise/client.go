package apprise

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Client represents an active apprise notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifApprise
	meta model.Meta
}

// New creates a new apprise notification instance
func New(config *model.NotifApprise, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "apprise"
}

// Send creates and sends a apprise notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
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
		Message string   `json:"body"`
		Title   string   `json:"title"`
		Tags    []string `json:"tags"`
		URLs    []string `json:"urls"`
	}{
		Message: string(body),
		Title:   string(title),
		Tags:    c.cfg.Tags,
		URLs:    c.cfg.URLs,
	}); err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "notify")

	if c.cfg.Token != "" || c.cfg.TokenFile != "" {
		token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
		if err != nil {
			return errors.Wrap(err, "cannot retrieve token secret for Apprise notifier")
		}
		u.Path = path.Join(u.Path, token)
	}

	q := u.Query()
	u.RawQuery = q.Encode()

	hc := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), *c.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), dataBuf)
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

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error            string `json:"error"`
			ErrorCode        int    `json:"errorCode"`
			ErrorDescription string `json:"errorDescription"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return errors.Wrapf(err, "cannot decode JSON error response for HTTP %d %s status", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return errors.Errorf("%d %s: %s", errBody.ErrorCode, errBody.Error, errBody.ErrorDescription)
	}

	return nil
}
