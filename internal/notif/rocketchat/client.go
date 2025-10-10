package rocketchat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/pkg/errors"
)

// Client represents an active rocketchat notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifRocketChat
	meta model.Meta
}

// New creates a new rocketchat notification instance
func New(config *model.NotifRocketChat, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "rocketchat"
}

// Send creates and sends a rocketchat notification with an entry
// https://rocket.chat/docs/developer-guides/rest-api/chat/postmessage/
func (c *Client) Send(entry model.NotifEntry) error {
	token, err := utl.GetValueOrFileContents(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for RocketChat notifier")
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

	title, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	var attachments []Attachment
	if *c.cfg.RenderAttachment {
		fields := []AttachmentField{
			{
				Title: "Hostname",
				Value: c.meta.Hostname,
				Short: false,
			},
			{
				Title: "Provider",
				Value: entry.Provider,
				Short: false,
			},
			{
				Title: "Created",
				Value: entry.Manifest.Created.Format("Jan 02, 2006 15:04:05 UTC"),
				Short: false,
			},
			{
				Title: "Digest",
				Value: entry.Manifest.Digest.String(),
				Short: false,
			},
			{
				Title: "Platform",
				Value: entry.Manifest.Platform,
				Short: false,
			},
		}
		if len(entry.Image.HubLink) > 0 {
			fields = append(fields, AttachmentField{
				Title: "HubLink",
				Value: entry.Image.HubLink,
				Short: false,
			})
		}
		attachments = append(attachments, Attachment{
			Text:   string(body),
			Ts:     json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
			Fields: fields,
		})
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(Message{
		Alias:       c.meta.Name,
		Avatar:      c.meta.Logo,
		Channel:     c.cfg.Channel,
		Text:        string(title),
		Attachments: attachments,
	}); err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api/v1/chat.postMessage")

	cancelCtx, cancel := context.WithCancelCause(context.Background())
	timeoutCtx, _ := context.WithTimeoutCause(cancelCtx, *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded)) //nolint:govet // no need to manually cancel this context as we already rely on parent
	defer func() { cancel(errors.WithStack(context.Canceled)) }()

	tlsConfig, err := utl.LoadTLSConfig(c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot load TLS configuration for Rocket.Chat notifier")
	}
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequestWithContext(timeoutCtx, "POST", u.String(), dataBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)
	req.Header.Add("X-User-Id", c.cfg.UserID)
	req.Header.Add("X-Auth-Token", token)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody struct {
		Success   bool   `json:"success"`
		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return errors.Wrapf(err, "cannot decode JSON body response for HTTP %d %s status", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected HTTP error %d: %s", resp.StatusCode, respBody.ErrorType)
	}
	return nil
}
