package pushover

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const pushoverAPIURL = "https://api.pushover.net/1/messages.json"

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
	token, err := utl.GetValueOrFileContents(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Pushover notifier")
	} else if token == "" {
		return errors.New("Pushover API token cannot be empty")
	}

	recipient, err := utl.GetValueOrFileContents(c.cfg.Recipient, c.cfg.RecipientFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve recipient secret for Pushover notifier")
	} else if recipient == "" {
		return errors.New("Pushover recipient cannot be empty")
	}

	message, err := msg.New(msg.Options{
		Meta:          c.meta,
		Entry:         entry,
		TemplateTitle: c.cfg.TemplateTitle,
		TemplateBody:  c.cfg.TemplateBody,
	})
	if err != nil {
		return err
	}

	title, body, err := message.RenderHTML()
	if err != nil {
		return err
	}

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	form := url.Values{}
	form.Add("token", token)
	form.Add("user", recipient)
	form.Add("title", string(title))
	form.Add("message", string(body))
	form.Add("priority", strconv.Itoa(c.cfg.Priority))
	if c.cfg.Sound != "" {
		form.Add("sound", c.cfg.Sound)
	}
	if c.meta.URL != "" {
		form.Add("url", c.meta.URL)
	}
	if c.meta.Name != "" {
		form.Add("url_title", c.meta.Name)
	}
	form.Add("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	form.Add("html", "1")

	hc := http.Client{}
	req, err := http.NewRequestWithContext(timeoutCtx, "POST", pushoverAPIURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.meta.UserAgent)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Header != nil {
		var appLimit, appRemaining int
		var appReset time.Time
		if limit := resp.Header.Get("X-Limit-App-Limit"); limit != "" {
			if i, err := strconv.Atoi(limit); err == nil {
				appLimit = i
			}
		}
		if remaining := resp.Header.Get("X-Limit-App-Remaining"); remaining != "" {
			if i, err := strconv.Atoi(remaining); err == nil {
				appRemaining = i
			}
		}
		if reset := resp.Header.Get("X-Limit-App-Reset"); reset != "" {
			if i, err := strconv.Atoi(reset); err == nil {
				appReset = time.Unix(int64(i), 0)
			}
		}
		log.Debug().Msgf("Pushover app limit: %d, remaining: %d, reset: %s", appLimit, appRemaining, appReset)
	}

	var respBody struct {
		Status  int      `json:"status"`
		Request string   `json:"request"`
		Errors  []string `json:"errors"`
		User    string   `json:"user"`
		Token   string   `json:"token"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return errors.Wrapf(err, "cannot decode JSON body response for HTTP %d %s status: %+v", resp.StatusCode, http.StatusText(resp.StatusCode), respBody)
	}
	if respBody.Status != 1 {
		return errors.Errorf("Pushover API call failed with status %d: %v", respBody.Status, respBody.Errors)
	}

	return nil
}
