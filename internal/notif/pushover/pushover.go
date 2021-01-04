package pushover

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/gregdel/pushover"
)

// Client represents an active Pushover notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifPushover
	meta model.Meta
}

// New creates a new Pushover notification instance
func New(config *model.NotifPushover, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "pushover"
}

// Send creates and sends a Pushover notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.New("Cannot retrieve token secret for Pushover notifier")
	}

	recipient, err := utl.GetSecret(c.cfg.Recipient, c.cfg.RecipientFile)
	if err != nil {
		return errors.New("Cannot retrieve recipient secret for Pushover notifier")
	}

	app := pushover.New(token)
	user := pushover.NewRecipient(recipient)

	title := fmt.Sprintf("Image update for %s", entry.Image.String())
	if entry.Status == model.ImageStatusNew {
		title = fmt.Sprintf("New image %s has been added", entry.Image.String())
	}

	tagTpl := "{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}"
	if len(entry.Image.HubLink) > 0 {
		tagTpl = `<a href="{{ .Entry.Image.HubLink }}">{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}</a>`
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("email").Parse(fmt.Sprintf("Docker tag %s which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status \"new\") }}newly added{{ else }}updated{{ end }} on {{ .Hostname }}.", tagTpl)))
	if err := msgTpl.Execute(&msgBuf, struct {
		Hostname string
		Entry    model.NotifEntry
	}{
		Hostname: c.meta.Hostname,
		Entry:    entry,
	}); err != nil {
		return err
	}

	_, err = app.GetRecipientDetails(user)
	if err != nil {
		return err
	}

	_, err = app.SendMessage(&pushover.Message{
		Message:   msgBuf.String(),
		Title:     title,
		Priority:  c.cfg.Priority,
		URL:       c.meta.URL,
		URLTitle:  c.meta.Name,
		Timestamp: time.Now().Unix(),
		HTML:      true,
	}, user)

	return err
}
