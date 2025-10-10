package gotify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Client represents an active gotify notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifGotify
	meta model.Meta
}

// New creates a new gotify notification instance
func New(config *model.NotifGotify, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "gotify"
}

// Send creates and sends a gotify notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	token, err := utl.GetValueOrFileContents(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Gotify notifier")
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

	jsonBody, err := json.Marshal(struct {
		Message  string                 `json:"message"`
		Title    string                 `json:"title"`
		Priority int                    `json:"priority"`
		Extras   map[string]interface{} `json:"extras"`
	}{
		Message:  string(body),
		Title:    string(title),
		Priority: c.cfg.Priority,
		Extras: map[string]interface{}{
			"client::display": map[string]string{
				"contentType": "text/markdown",
			},
		},
	})
	if err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "message")

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	tlsConfig, err := utl.LoadTLSConfig(c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot load TLS configuration for Gotify notifier")
	}
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	req, err := http.NewRequestWithContext(timeoutCtx, "POST", u.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(string(jsonBody))))
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
