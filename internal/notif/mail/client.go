package mail

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	hermes "github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/wneessen/go-mail"
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

	email := hermes.Email{
		Body: hermes.Body{
			Title:        fmt.Sprintf("%s 🔔 notification", c.meta.Name),
			FreeMarkdown: hermes.Markdown(body),
			Signature:    "Thanks for your support!",
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	htmlpart, err := h.GenerateHTML(email)
	if err != nil {
		return errors.Wrap(err, "cannot generate HTML email")
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	textpart, err := h.GeneratePlainText(email)
	if err != nil {
		return errors.Wrap(err, "cannot generate plaintext email")
	}

	mailMessage := mail.NewMsg()
	if err = mailMessage.FromFormat(c.meta.Name, c.cfg.From); err != nil {
		return errors.Wrap(err, "cannot set mail FROM address")
	}
	if err = mailMessage.To(c.cfg.To...); err != nil {
		return errors.Wrap(err, "cannot set mail TO address(es)")
	}
	mailMessage.Subject(string(title))
	mailMessage.SetBodyString(mail.TypeTextPlain, textpart)
	mailMessage.AddAlternativeString(mail.TypeTextHTML, htmlpart)

	username, err := utl.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for mail notifier")
	}
	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for mail notifier")
	}
	localname := c.cfg.LocalName
	if localname == "" {
		c.cfg.LocalName, err = os.Hostname()
		if err != nil {
			log.Warn().Err(err).Msg("Cannot retrieve hostname for local name")
		}
	}

	client, err := mail.NewClient(c.cfg.Host,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithPort(c.cfg.Port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithHELO(localname),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot create mail client")
	}
	if *c.cfg.SSL {
		client.SetSSL(*c.cfg.SSL)
	}
	if *c.cfg.InsecureSkipVerify {
		if err = client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}); err != nil {
			log.Warn().Err(err).Msg("Cannot set TLS config")
		}
	}

	return client.DialAndSend(mailMessage)
}
