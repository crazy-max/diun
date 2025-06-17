package telegram

import (
	"encoding/json"
	"net/http"
	"strconv"
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

type chatID struct {
	id     int64
	topics []int64
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

	cids := c.cfg.ChatIDs
	cidsRaw, err := utl.GetSecret("", c.cfg.ChatIDsFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve chat IDs secret for Telegram notifier")
	}
	if len(cidsRaw) > 0 {
		if err = json.Unmarshal([]byte(cidsRaw), &cids); err != nil {
			return errors.Wrap(err, "cannot unmarshal chat IDs secret for Telegram notifier")
		}
	}
	if len(cids) == 0 {
		return errors.New("no chat IDs provided for Telegram notifier")
	}

	parsedChatIDs, err := parseChatIDs(cids)
	if err != nil {
		return errors.Wrap(err, "cannot parse chat IDs for Telegram notifier")
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

	disableNotification := false
	if c.cfg.DisableNotification != nil {
		disableNotification = *c.cfg.DisableNotification
	}

	for _, cid := range parsedChatIDs {
		if len(cid.topics) > 0 {
			for _, topic := range cid.topics {
				if err = sendTelegramMessage(bot, cid.id, topic, string(body), disableNotification); err != nil {
					return err
				}
			}
		} else {
			if err = sendTelegramMessage(bot, cid.id, 0, string(body), disableNotification); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseChatIDs(entries []string) ([]chatID, error) {
	var chatIDs []chatID
	for _, entry := range entries {
		parts := strings.Split(entry, ":")
		if len(parts) < 1 || len(parts) > 2 {
			return nil, errors.Errorf("invalid chat ID %q", entry)
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "invalid chat ID")
		}
		var topics []int64
		if len(parts) == 2 {
			topicParts := strings.Split(parts[1], ";")
			for _, topicPart := range topicParts {
				topic, err := strconv.ParseInt(topicPart, 10, 64)
				if err != nil {
					return nil, errors.Wrapf(err, "invalid topic %q for chat ID %d", topicPart, id)
				}
				topics = append(topics, topic)
			}
		}
		chatIDs = append(chatIDs, chatID{
			id:     id,
			topics: topics,
		})
	}
	return chatIDs, nil
}

func sendTelegramMessage(bot *gotgbot.Bot, chatID int64, threadID int64, message string, disableNotification bool) error {
	_, err := bot.SendMessage(chatID, message, &gotgbot.SendMessageOpts{
		MessageThreadId:     threadID,
		ParseMode:           gotgbot.ParseModeMarkdown,
		LinkPreviewOptions:  &gotgbot.LinkPreviewOptions{IsDisabled: true},
		DisableNotification: disableNotification,
	})
	return err
}
