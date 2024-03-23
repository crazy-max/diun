package telegram

import (
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
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
	token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Telegram notifier")
	}

	chatIDs := c.cfg.ChatIDs
	chatIDsRaw, err := utl.GetSecret("", c.cfg.ChatIDsFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve chat IDs secret for Telegram notifier")
	}
	if len(chatIDsRaw) > 0 {
		if err = json.Unmarshal([]byte(chatIDsRaw), &chatIDs); err != nil {
			return errors.Wrap(err, "cannot unmarshal chat IDs secret for Telegram notifier")
		}
	}

	bot, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: http.Client{},
			DefaultRequestOpts: &gotgbot.RequestOpts{
				Timeout: gotgbot.DefaultTimeout,
				APIURL:  gotgbot.DefaultAPIURL,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create telegram bot client")
	}

	message, err := msg.New(msg.Options{
		Meta:         c.meta,
		Entry:        entry,
		TemplateBody: c.cfg.TemplateBody,
		TemplateFuncs: template.FuncMap{
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

	_, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	for _, chatID := range chatIDs {
		_, err := bot.SendMessage(chatID, string(body), &gotgbot.SendMessageOpts{
			ParseMode:          gotgbot.ParseModeMarkdown,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
