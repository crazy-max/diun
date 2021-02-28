package rocketchat

import (
	"bytes"
	"encoding/json"
	"fmt"
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

const customTpl = `Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }} on {{ .Meta.Hostname }}.`

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
	token, err := utl.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.New("Cannot retrieve token secret for RocketChat notifier")
	}

	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	title, text, err := message.RenderMarkdownTemplate(customTpl)
	if err != nil {
		return err
	}

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

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(Message{
		Alias:   c.meta.Name,
		Avatar:  c.meta.Logo,
		Channel: c.cfg.Channel,
		Text:    title,
		Attachments: []Attachment{
			{
				Text:   string(text),
				Ts:     json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
				Fields: fields,
			},
		},
	}); err != nil {
		return err
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api/v1/chat.postMessage")

	req, err := http.NewRequest("POST", u.String(), dataBuf)
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

	var respBody struct {
		Success   bool   `json:"success"`
		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err == nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, respBody.ErrorType)
	}

	return nil
}
