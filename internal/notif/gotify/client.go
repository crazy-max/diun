package gotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
)

// Client represents an active gotify notification object
type Client struct {
	*notifier.Notifier
	cfg model.NotifGotify
	app model.App
}

// New creates a new gotify notification instance
func New(config model.NotifGotify, app model.App) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
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
		Timeout: time.Duration(c.cfg.Timeout) * time.Second,
	}

	title := fmt.Sprintf("Image update for %s", entry.Image.String())
	if entry.Status == model.ImageStatusNew {
		title = fmt.Sprintf("New image %s has been added", entry.Image.String())
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("gotify").Parse(`Docker 🐳 tag {{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }} which you subscribed to through {{ .Provider }} provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.`))
	if err := msgTpl.Execute(&msgBuf, entry); err != nil {
		return err
	}

	data := url.Values{}
	data.Set("message", msgBuf.String())
	data.Set("title", title)
	data.Set("priority", strconv.Itoa(c.cfg.Priority))

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "message")

	q := u.Query()
	q.Set("token", c.cfg.Token)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Set("User-Agent", fmt.Sprintf("%s %s", c.app.Name, c.app.Version))

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
