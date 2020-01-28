package telegram

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog/log"
)

// Client represents an active Telegram notification object
type Client struct {
	*notifier.Notifier
	cfg model.NotifTelegram
	app model.App
	bot *tgbotapi.BotAPI
}

// New creates a new Telegram notification instance
func New(config model.NotifTelegram, app model.App) notifier.Notifier {
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Err(err).Msgf("Failed to initialize Telegram notifications")
	}
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
			bot: bot,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "telegram"
}

// Send creates and sends a Telegram notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	if c.bot == nil {
		return errors.New("Telegram not initialized")
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("email").Parse(`Docker 🐳 tag {{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }} which you subscribed to through {{ .Provider }} provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.`))
	if err := msgTpl.Execute(&msgBuf, entry); err != nil {
		return err
	}

	for _, chatID := range c.cfg.ChatIDs {
		msg := tgbotapi.NewMessage(chatID, c.bot.Self.UserName)
		msg.Text = msgBuf.String()
		if _, err := c.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}
