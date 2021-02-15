package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/microcosm-cc/bluemonday"
	"github.com/opencontainers/go-digest"
	"github.com/russross/blackfriday/v2"
)

// Client represents an active msg object
type Client struct {
	opts Options
}

// Options holds msg client object options
type Options struct {
	Meta     model.Meta
	Entry    model.NotifEntry
	TplFuncs template.FuncMap
}

const defaultTpl = `Docker tag {{ if .Entry.Image.HubLink }}[**{{ .Entry.Image }}**]({{ .Entry.Image.HubLink }}){{ else }}**{{ .Entry.Image }}**{{ end }}
which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}
on {{ .Meta.Hostname }}.`

// New initializes a new msg client
func New(opts Options) (*Client, error) {
	return &Client{
		opts,
	}, nil
}

// RenderMarkdown returns a notification message as markdown
func (c *Client) RenderMarkdown() (title string, text []byte, err error) {
	return c.RenderMarkdownTemplate(strings.ReplaceAll(defaultTpl, "\n", " "))
}

// RenderMarkdownTemplate returns a notification message as markdown with a custom template
func (c *Client) RenderMarkdownTemplate(tpl string) (title string, text []byte, err error) {
	title = fmt.Sprintf("Image update for %s", c.opts.Entry.Image.String())
	if c.opts.Entry.Status == model.ImageStatusNew {
		title = fmt.Sprintf("New image %s has been added", c.opts.Entry.Image.String())
	}

	var msgBuf bytes.Buffer
	msgTpl := template.Must(template.New("notif").Funcs(c.opts.TplFuncs).Parse(tpl))
	err = msgTpl.Execute(&msgBuf, struct {
		Meta  model.Meta
		Entry model.NotifEntry
	}{
		Meta:  c.opts.Meta,
		Entry: c.opts.Entry,
	})

	text = msgBuf.Bytes()
	return
}

// RenderHTML returns a notification message as html
func (c *Client) RenderHTML() (title string, text []byte, err error) {
	return c.RenderHTMLTemplate(strings.ReplaceAll(defaultTpl, "\n", " "))
}

// RenderHTMLTemplate returns a notification message as html with a custom template
func (c *Client) RenderHTMLTemplate(tpl string) (title string, text []byte, err error) {
	title, text, err = c.RenderMarkdownTemplate(tpl)
	if err != nil {
		return
	}

	text = []byte(bluemonday.UGCPolicy().Sanitize(
		// Dirty way to remove wrapped <p></p> and newline
		// https://github.com/russross/blackfriday/issues/237
		strings.TrimRight(strings.TrimLeft(strings.TrimSpace(string(blackfriday.Run(text))), "<p>"), "</p>"),
	))
	return
}

// RenderJSON returns a notification message as JSON
func (c *Client) RenderJSON() ([]byte, error) {
	return json.Marshal(struct {
		Version  string        `json:"diun_version"`
		Hostname string        `json:"hostname"`
		Status   string        `json:"status"`
		Provider string        `json:"provider"`
		Image    string        `json:"image"`
		HubLink  string        `json:"hub_link"`
		MIMEType string        `json:"mime_type"`
		Digest   digest.Digest `json:"digest"`
		Created  *time.Time    `json:"created"`
		Platform string        `json:"platform"`
	}{
		Version:  c.opts.Meta.Version,
		Hostname: c.opts.Meta.Hostname,
		Status:   string(c.opts.Entry.Status),
		Provider: c.opts.Entry.Provider,
		Image:    c.opts.Entry.Image.String(),
		HubLink:  c.opts.Entry.Image.HubLink,
		MIMEType: c.opts.Entry.Manifest.MIMEType,
		Digest:   c.opts.Entry.Manifest.Digest,
		Created:  c.opts.Entry.Manifest.Created,
		Platform: c.opts.Entry.Manifest.Platform,
	})
}

// RenderEnv returns a notification message as environment variables
func (c *Client) RenderEnv() []string {
	return []string{
		fmt.Sprintf("DIUN_VERSION=%s", c.opts.Meta.Version),
		fmt.Sprintf("DIUN_HOSTNAME=%s", c.opts.Meta.Hostname),
		fmt.Sprintf("DIUN_ENTRY_STATUS=%s", string(c.opts.Entry.Status)),
		fmt.Sprintf("DIUN_ENTRY_PROVIDER=%s", c.opts.Entry.Provider),
		fmt.Sprintf("DIUN_ENTRY_IMAGE=%s", c.opts.Entry.Image.String()),
		fmt.Sprintf("DIUN_ENTRY_HUBLINK=%s", c.opts.Entry.Image.HubLink),
		fmt.Sprintf("DIUN_ENTRY_MIMETYPE=%s", c.opts.Entry.Manifest.MIMEType),
		fmt.Sprintf("DIUN_ENTRY_DIGEST=%s", c.opts.Entry.Manifest.Digest),
		fmt.Sprintf("DIUN_ENTRY_CREATED=%s", c.opts.Entry.Manifest.Created),
		fmt.Sprintf("DIUN_ENTRY_PLATFORM=%s", c.opts.Entry.Manifest.Platform),
	}
}
