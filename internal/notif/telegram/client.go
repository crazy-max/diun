package telegram

import (
	"strings"
	"text/template"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Client represents an active Telegram notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifTelegram
	meta model.Meta
}

const customTpl = `Docker tag {{ if .Entry.Image.HubLink }}[{{ .Entry.Image }}]({{ .Entry.Image.HubLink }}){{ else }}{{ .Entry.Image }}{{ end }}
which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}
on {{ escapeMarkdown .Meta.Hostname }}.`

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

	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
		TplFuncs: template.FuncMap{
			"escapeMarkdown": func(text string) string {
				text = strings.ReplaceAll(text, "_", "\\_")
				text = strings.ReplaceAll(text, "*", "\\*")
				text = strings.ReplaceAll(text, "[", "\\[")
				text = strings.ReplaceAll(text, "`", "\\`")
				return text
			},
		},
	})
	if err != nil {
		return err
	}

	_, text, err := message.RenderMarkdownTemplate(strings.ReplaceAll(customTpl, "\n", " "))
	if err != nil {
		return err
	}

	for _, chatID := range c.cfg.ChatIDs {
		_, err := bot.Send(tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: chatID,
			},
			Text:                  string(text),
			ParseMode:             "markdown",
			DisableWebPagePreview: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
