package telegram

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Client represents an active Telegram notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifTelegram
	meta model.Meta
}

// New creates a new Telegram notification instance
func New(config *model.NotifTelegram, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "telegram"
}

// Send creates and sends a Telegram notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	bot, err := tgbotapi.NewBotAPI(c.cfg.Token)
	if err != nil {
		return err
	}

	tagTpl := "{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}"
	if len(entry.Image.HubLink) > 0 {
		tagTpl = "[{{ .Entry.Image.Domain }}/{{ .Entry.Image.Path }}:{{ .Entry.Image.Tag }}]({{ .Entry.Image.HubLink }})"
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("email").Parse(fmt.Sprintf("Docker tag %s which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status \"new\") }}newly added{{ else }}updated{{ end }} on {{ .Meta.Hostname }}.", tagTpl)))
	if err := msgTpl.Execute(&msgBuf, struct {
		Meta  model.Meta
		Entry model.NotifEntry
	}{
		Meta:  c.meta,
		Entry: entry,
	}); err != nil {
		return err
	}

	for _, chatID := range c.cfg.ChatIDs {
		_, err := bot.Send(tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: chatID,
			},
			Text:                  msgBuf.String(),
			ParseMode:             "markdown",
			DisableWebPagePreview: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
