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
	"github.com/pkg/errors"
	"github.com/russross/blackfriday/v2"
)

// Client represents an active msg object
type Client struct {
	opts Options
}

// Options holds msg client object options
type Options struct {
	Meta          model.Meta
	Entry         model.NotifEntry
	TemplateTitle string
	TemplateBody  string
	TemplateFuncs template.FuncMap
}

// New initializes a new msg client
func New(opts Options) (*Client, error) {
	return &Client{
		opts,
	}, nil
}

// RenderMarkdown returns a notification message as markdown
func (c *Client) RenderMarkdown() (title []byte, body []byte, _ error) {
	var err error

	title, err = c.RenderTemplate("title", c.opts.TemplateTitle)
	if err != nil {
		return title, body, err
	}

	body, err = c.RenderTemplate("body", c.opts.TemplateBody)
	if err != nil {
		return title, body, err
	}

	return
}

// RenderTemplate renders a notification template with the entry context.
func (c *Client) RenderTemplate(name, text string) ([]byte, error) {
	var buf bytes.Buffer
	tpl, err := template.New(name).Funcs(templateFuncs(c.opts.TemplateFuncs)).Parse(strings.TrimSuffix(strings.TrimSpace(text), "\n"))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse %s template", name)
	}
	if err = tpl.Execute(&buf, struct {
		Meta  model.Meta
		Entry model.NotifEntry
	}{
		Meta:  c.opts.Meta,
		Entry: c.opts.Entry,
	}); err != nil {
		return nil, errors.Wrapf(err, "cannot render notif %s", name)
	}

	return buf.Bytes(), nil
}

// RenderHTML returns a notification message as html
func (c *Client) RenderHTML() (title []byte, body []byte, err error) {
	title, body, err = c.RenderMarkdown()
	if err != nil {
		return title, body, err
	}

	htmlBody := strings.TrimSpace(string(blackfriday.Run(body)))
	// Dirty way to remove wrapped <p></p> and newline
	// https://github.com/russross/blackfriday/issues/237
	htmlBody = strings.TrimPrefix(htmlBody, "<p>")
	htmlBody = strings.TrimSuffix(htmlBody, "</p>")
	body = []byte(bluemonday.UGCPolicy().Sanitize(htmlBody))
	return
}

// RenderJSON returns a notification message as JSON
func (c *Client) RenderJSON() ([]byte, error) {
	return json.Marshal(struct {
		Version  string            `json:"diun_version"`
		Hostname string            `json:"hostname"`
		Status   string            `json:"status"`
		Provider string            `json:"provider"`
		Image    string            `json:"image"`
		HubLink  string            `json:"hub_link"`
		MIMEType string            `json:"mime_type"`
		Digest   digest.Digest     `json:"digest"`
		Created  *time.Time        `json:"created"`
		Platform string            `json:"platform"`
		Metadata map[string]string `json:"metadata"`
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
		Metadata: c.opts.Entry.Metadata,
	})
}

// RenderEnv returns a notification message as environment variables
func (c *Client) RenderEnv() []string {
	var metadataEnvs []string
	for k, v := range c.opts.Entry.Metadata {
		metadataEnvs = append(metadataEnvs, fmt.Sprintf("DIUN_ENTRY_METADATA_%s=%s", strings.ToUpper(k), v))
	}
	return append([]string{
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
	}, metadataEnvs...)
}
