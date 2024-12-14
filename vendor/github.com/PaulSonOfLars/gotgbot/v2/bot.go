package gotgbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

//go:generate go run ./scripts/generate

// Bot is the default Bot struct used to send and receive messages to the telegram API.
type Bot struct {
	// Token stores the bot's secret token obtained from t.me/BotFather, and used to interact with telegram's API.
	Token string

	// The bot's User info, as returned by Bot.GetMe. Populated when created through the NewBot method.
	User
	// The bot client to use to make requests
	BotClient
}

// BotOpts declares all optional parameters for the NewBot function.
type BotOpts struct {
	// BotClient allows for passing in custom configurations of BotClient, such as handling extra errors or providing
	// metrics.
	BotClient BotClient
	// DisableTokenCheck can be used to disable the token validity check.
	// Useful when running in time-constrained environments where the startup time should be minimised, and where the
	// token can be assumed to be valid (eg lambdas).
	// Warning: Disabling the token check will mean that the Bot.User struct will no longer be populated.
	DisableTokenCheck bool
	// Request opts to use for checking token validity with Bot.GetMe. Can be slow - a high timeout (eg 10s) is
	// recommended.
	RequestOpts *RequestOpts
}

// NewBot returns a new Bot struct populated with the necessary defaults.
func NewBot(token string, opts *BotOpts) (*Bot, error) {
	botClient := BotClient(&BaseBotClient{
		Client:             http.Client{},
		UseTestEnvironment: false,
		DefaultRequestOpts: nil,
	})

	// Large timeout on the initial GetMe request as this can sometimes be slow.
	getMeReqOpts := &RequestOpts{
		Timeout: 10 * time.Second,
	}

	checkTokenValidity := true
	if opts != nil {
		if opts.BotClient != nil {
			botClient = opts.BotClient
		}

		if opts.RequestOpts != nil {
			getMeReqOpts = opts.RequestOpts
		}
		checkTokenValidity = !opts.DisableTokenCheck
	}

	b := Bot{
		Token:     token,
		BotClient: botClient,
	}

	if checkTokenValidity {
		// Get bot info. This serves two purposes:
		// 1. Check token is valid.
		// 2. Populate the bot struct "User" field.
		botUser, err := b.GetMe(&GetMeOpts{RequestOpts: getMeReqOpts})
		if err != nil {
			return nil, fmt.Errorf("failed to check bot token: %w", err)
		}
		b.User = *botUser
	}

	return &b, nil
}

// UseMiddleware allows you to wrap the existing bot client to enhance functionality
//
// Deprecated: Instead of using middlewares, consider implementing the BotClient interface.
func (bot *Bot) UseMiddleware(mw func(client BotClient) BotClient) *Bot {
	bot.BotClient = mw(bot.BotClient)
	return bot
}

var ErrNilBotClient = errors.New("nil BotClient")

func (bot *Bot) Request(method string, params map[string]string, data map[string]NamedReader, opts *RequestOpts) (json.RawMessage, error) {
	if bot.BotClient == nil {
		return nil, ErrNilBotClient
	}

	ctx, cancel := bot.BotClient.TimeoutContext(opts)
	defer cancel()

	return bot.BotClient.RequestWithContext(ctx, bot.Token, method, params, data, opts)
}
