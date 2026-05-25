package mail

import (
	"crypto/tls"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/internal/secret"
	hermes "github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	email "github.com/wneessen/go-mail"
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

	message, err := msg.New(msg.Options{
		Meta:          c.meta,
		Entry:         entry,
		TemplateTitle: c.cfg.TemplateTitle,
		TemplateBody:  c.cfg.TemplateBody,
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

	title, body, err := message.RenderMarkdown()
	if err != nil {
		return err
	}

	hermesEmail := hermes.Email{
		Body: hermes.Body{
			Title:        fmt.Sprintf("%s 🔔 notification", c.meta.Name),
			FreeMarkdown: hermes.Markdown(body),
			Signature:    "Thanks for your support!",
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	htmlpart, err := h.GenerateHTML(hermesEmail)
	if err != nil {
		return errors.Wrap(err, "cannot generate HTML email")
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	textpart, err := h.GeneratePlainText(hermesEmail)
	if err != nil {
		return errors.Wrap(err, "cannot generate plaintext email")
	}

	mailMessage := email.NewMsg()
	if err = mailMessage.FromFormat(c.meta.Name, c.cfg.From); err != nil {
		return errors.Wrap(err, "cannot set mail FROM address")
	}
	if err = mailMessage.To(c.cfg.To...); err != nil {
		return errors.Wrap(err, "cannot set mail TO address(es)")
	}
	mailMessage.Subject(string(title))
	mailMessage.SetBodyString(email.TypeTextPlain, textpart)
	mailMessage.AddAlternativeString(email.TypeTextHTML, htmlpart)

	username, err := secret.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for mail notifier")
	}
	password, err := secret.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for mail notifier")
	}

	client, err := c.mailClient(username, password)
	if err != nil {
		return errors.Wrap(err, "cannot create mail client")
	}

	if err = client.DialAndSend(mailMessage); err != nil {
		return errors.Wrap(err, "cannot send mail notification")
	}
	return nil
}

func (c *Client) mailClient(username, password string) (*email.Client, error) {
	localName := c.cfg.LocalName
	if localName == "" {
		localName = "localhost"
	}
	opts := []email.Option{
		email.WithPort(c.cfg.Port),
		email.WithTLSPolicy(email.TLSOpportunistic),
		email.WithHELO(localName),
	}
	if *c.cfg.SSL {
		opts = append(opts, email.WithSSL())
	}
	if *c.cfg.InsecureSkipVerify {
		opts = append(opts, email.WithTLSConfig(&tls.Config{
			ServerName:         c.cfg.Host,
			MinVersion:         email.DefaultTLSMinVersion,
			InsecureSkipVerify: true,
		}))
	}
	if username != "" {
		opts = append(opts,
			email.WithUsername(username),
			email.WithPassword(password),
			email.WithSMTPAuth(email.SMTPAuthAutoDiscover),
		)
	}
	return email.NewClient(c.cfg.Host, opts...)
}
