package gotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
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
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	title := fmt.Sprintf("Image update for %s", entry.Image.String())
	if entry.Status == model.ImageStatusNew {
		title = fmt.Sprintf("New image %s has been added", entry.Image.String())
	}

	tagTpl := "`{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}`"
	if len(entry.Image.HubLink) > 0 {
		tagTpl = "[`{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}`]({{ .Entry.Image.HubLink }})"
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("gotify").Parse(fmt.Sprintf("Docker tag %s which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status \"new\") }}newly added{{ else }}updated{{ end }} on {{ .Meta.Hostname }}.", tagTpl)))
	if err := msgTpl.Execute(&msgBuf, struct {
		Meta  model.Meta
		Entry model.NotifEntry
	}{
		Meta:  c.meta,
		Entry: entry,
	}); err != nil {
		return err
	}

	var body, err = json.Marshal(struct {
		Message  string                 `json:"message"`
		Title    string                 `json:"title"`
		Priority int                    `json:"priority"`
		Extras   map[string]interface{} `json:"extras"`
	}{
		Message:  msgBuf.String(),
		Title:    title,
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
	q.Set("token", c.cfg.Token)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(string(body))))
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
		return fmt.Errorf("%d %s: %s", errBody.ErrorCode, errBody.Error, errBody.ErrorDescription)
	}

	return nil
}
