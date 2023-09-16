package gotify

import (
	"bytes"
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
	token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Gotify notifier")
	}

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

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonBody))
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
