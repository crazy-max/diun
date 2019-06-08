package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"text/template"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/go-gomail/gomail"
	"github.com/matcornic/hermes/v2"
)

// Client represents an active mail notification object
type Client struct {
	*notifier.Notifier
	cfg model.Mail
	app model.App
}

// New creates a new mail notification instance
func New(config model.Mail, app model.App) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
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
			Name: c.app.Name,
			Link: "https://github.com/crazy-max/diun",
			Logo: "https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png",
			Copyright: fmt.Sprintf("%s © %d %s %s",
				c.app.Author,
				time.Now().Year(),
				c.app.Name,
				c.app.Version),
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

Docker 🐳 tag **{{ .Image.Domain }}/{{ .Image.Path }}:{{ .Image.Tag }}** which you subscribed to has been {{ if (eq .Status "new") }}newly added{{ else }}updated{{ end }}.

This image has been {{ if (eq .Status "new") }}created{{ else }}updated{{ end }} at <code>{{ .Manifest.Created }}</code> with digest <code>{{ .Manifest.Digest }}</code> for <code>{{ .Manifest.Os }}/{{ .Manifest.Architecture }}</code> platform.

Need help, or have questions? Go to https://github.com/crazy-max/diun and leave an issue.

`))
	if err := emailTpl.Execute(&emailBuf, entry); err != nil {
		return err
	}
	email := hermes.Email{
		Body: hermes.Body{
			Title:        fmt.Sprintf("%s 🔔 notification", c.app.Name),
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
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", c.app.Name, c.cfg.From))
	msg.SetHeader("To", c.cfg.To)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	var tlsConfig *tls.Config
	if c.cfg.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: c.cfg.InsecureSkipVerify,
		}
	}

	dialer := &gomail.Dialer{
		Host:      c.cfg.Host,
		Port:      c.cfg.Port,
		Username:  c.cfg.Username,
		Password:  c.cfg.Password,
		SSL:       c.cfg.SSL,
		TLSConfig: tlsConfig,
	}

	return dialer.DialAndSend(msg)
}
