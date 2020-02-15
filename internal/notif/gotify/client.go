package gotify

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
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
	msgTpl := template.Must(template.New("email").Parse(`Docker 🐳 tag {{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }} which you subscribed to through {{ .Provider }} provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.`))
	if err := msgTpl.Execute(&msgBuf, entry); err != nil {
		return err
	}

	_, err := hc.PostForm(c.cfg.Host+"/message?token="+c.cfg.Token,
		url.Values{"message": {msgBuf.String()}, "title": {title}})

	return err
}
