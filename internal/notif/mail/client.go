package mail

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/go-gomail/gomail"
	hermes "github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
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

// generateMessageID creates a unique Message-ID header value according to RFC 5322.
// Format: <timestamp.randomhex@domain>
// The domain is taken from the LocalName config if set, otherwise falls back to Host.
func (c *Client) generateMessageID() (string, error) {
	// Use nanosecond timestamp for high resolution uniqueness
	timestamp := time.Now().UnixNano()

	// Generate 8 bytes of cryptographically secure random data
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.Wrap(err, "failed to generate random bytes for Message-ID")
	}
	randomHex := hex.EncodeToString(randomBytes)

	// Use LocalName if set, otherwise fall back to Host
	domain := c.cfg.LocalName
	if domain == "" {
		domain = c.cfg.Host
	}

	// RFC 5322 format: <local-part@domain>
	messageID := fmt.Sprintf("<%d.%s@%s>", timestamp, randomHex, domain)
	return messageID, nil
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

	// Generate RFC 5322 compliant Message-ID
	messageID, err := c.generateMessageID()
	if err != nil {
		return errors.Wrap(err, "cannot generate Message-ID")
	}

	mailMessage := gomail.NewMessage()
	mailMessage.SetHeader("From", fmt.Sprintf("%s <%s>", c.meta.Name, c.cfg.From))
	mailMessage.SetHeader("To", c.cfg.To...)
	mailMessage.SetHeader("Subject", string(title))
	mailMessage.SetHeader("Message-ID", messageID)
	mailMessage.SetBody("text/plain", textpart)
	mailMessage.AddAlternative("text/html", htmlpart)

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
		LocalName: c.cfg.LocalName,
	}

	return dialer.DialAndSend(mailMessage)
}
