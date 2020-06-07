package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/opencontainers/go-digest"
	"github.com/streadway/amqp"
)

// Client represents an active amqp notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifAmqp
	meta model.Meta
}

// New creates a new amqp notification instance
func New(config *model.NotifAmqp, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "amqp"
}

// Send creates and sends a amqp notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	username, err := utl.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		return err
	}

	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		return err
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, c.cfg.Host, c.cfg.Port))
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		c.cfg.Queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(struct {
		Version  string        `json:"diun_version"`
		Status   string        `json:"status"`
		Provider string        `json:"provider"`
		Image    string        `json:"image"`
		MIMEType string        `json:"mime_type"`
		Digest   digest.Digest `json:"digest"`
		Created  *time.Time    `json:"created"`
		Platform string        `json:"platform"`
	}{
		Version:  c.meta.Version,
		Status:   string(entry.Status),
		Provider: entry.Provider,
		Image:    entry.Image.String(),
		MIMEType: entry.Manifest.MIMEType,
		Digest:   entry.Manifest.Digest,
		Created:  entry.Manifest.Created,
		Platform: entry.Manifest.Platform,
	})
	if err != nil {
		return err
	}

	return ch.Publish(
		c.cfg.Exchange,
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
