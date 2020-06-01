package telegram

import (
	"bytes"
	"text/template"

	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/internal/notif/notifier"
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

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("email").Parse(`Docker 🐳 tag {{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }} which you subscribed to through {{ .Provider }} provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.`))
	if err := msgTpl.Execute(&msgBuf, entry); err != nil {
		return err
	}

	for _, chatID := range c.cfg.ChatIDs {
		msg := tgbotapi.NewMessage(chatID, bot.Self.UserName)
		msg.Text = msgBuf.String()
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}
