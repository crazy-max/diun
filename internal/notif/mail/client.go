package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/go-gomail/gomail"
	"github.com/matcornic/hermes/v2"
	"github.com/rs/zerolog/log"
)

// Client represents an active mail notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifMail
	meta model.Meta
}

// New creates a new mail notification instance
func New(config *model.NotifMail, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "mail"
}

// Send creates and sends an email notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	h := hermes.Hermes{
		Theme: new(Theme),
		Product: hermes.Product{
			Name: c.meta.Name,
			Link: c.meta.URL,
			Logo: c.meta.Logo,
			Copyright: fmt.Sprintf("%s © %d %s %s",
				c.meta.Author,
				time.Now().Year(),
				c.meta.Name,
				c.meta.Version),
		},
	}

	// Subject
	subject := fmt.Sprintf("Image update for %s", entry.Image.String())
	if entry.Status == model.ImageStatusNew {
		subject = fmt.Sprintf("New image %s has been added", entry.Image.String())
	}

	// Body
	var emailBuf bytes.Buffer
	emailTpl := template.Must(template.New("email").Parse(`

Docker 🐳 tag **{{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }}** which you subscribed to through **{{ .Provider }}** provider has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.

This image has been {{ if (eq .Status "new") }}created{{ else }}updated{{ end }} at <code>{{ .Manifest.Created.Format "Jan 02, 2006 15:04:05 UTC" }}</code> with digest <code>{{ .Manifest.Digest }}</code> for <code>{{ .Manifest.Platform }}</code> platform.

Need help, or have questions? Go to https://github.com/crazy-max/diun and leave an issue.

`))
	if err := emailTpl.Execute(&emailBuf, entry); err != nil {
		return err
	}
	email := hermes.Email{
		Body: hermes.Body{
			Title:        fmt.Sprintf("%s 🔔 notification", c.meta.Name),
			FreeMarkdown: hermes.Markdown(emailBuf.String()),
			Signature:    "Thanks for your support",
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	htmlpart, err := h.GenerateHTML(email)
	if err != nil {
		return fmt.Errorf("hermes: %v", err)
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	textpart, err := h.GeneratePlainText(email)
	if err != nil {
		return fmt.Errorf("hermes: %v", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", c.meta.Name, c.cfg.From))
	msg.SetHeader("To", c.cfg.To)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	var tlsConfig *tls.Config
	if *c.cfg.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: *c.cfg.InsecureSkipVerify,
		}
	}

	username, err := utl.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for mail notifier")
	}
	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for mail notifier")
	}

	dialer := &gomail.Dialer{
		Host:      c.cfg.Host,
		Port:      c.cfg.Port,
		Username:  username,
		Password:  password,
		SSL:       *c.cfg.SSL,
		TLSConfig: tlsConfig,
	}

	return dialer.DialAndSend(msg)
}
