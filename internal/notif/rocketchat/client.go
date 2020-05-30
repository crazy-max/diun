package rocketchat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/internal/notif/notifier"
)

// Client represents an active rocketchat notification object
type Client struct {
	*notifier.Notifier
	cfg       *model.NotifRocketChat
	app       model.App
	userAgent string
}

// New creates a new rocketchat notification instance
func New(config *model.NotifRocketChat, app model.App, userAgent string) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:       config,
			app:       app,
			userAgent: userAgent,
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
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	title := fmt.Sprintf("Image update for %s", entry.Image.String())
	if entry.Status == model.ImageStatusNew {
		title = fmt.Sprintf("New image %s has been added", entry.Image.String())
	}

	var textBuf bytes.Buffer
	textTpl := template.Must(template.New("rocketchat").Parse(`Docker 🐳 tag {{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }} which you subscribed to through {{ .Provider }} provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.`))
	if err := textTpl.Execute(&textBuf, entry); err != nil {
		return err
	}

	data := Message{
		Alias:   c.app.Name,
		Avatar:  "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
		Channel: c.cfg.Channel,
		Text:    title,
		Attachments: []Attachment{
			{
				Text: textBuf.String(),
				Ts:   json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
				Fields: []AttachmentField{
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
				},
			},
		},
	}

	dataBuf := new(bytes.Buffer)
	if err := json.NewEncoder(dataBuf).Encode(data); err != nil {
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
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Add("X-User-Id", c.cfg.UserID)
	req.Header.Add("X-Auth-Token", c.cfg.Token)

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
