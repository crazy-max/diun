package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/notif/notifier"
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/opencontainers/go-digest"
	"github.com/streadway/amqp"
)

// Client represents an active amqp notification object
type Client struct {
	*notifier.Notifier
	cfg *model.NotifAmqp
	app model.App
}

// New creates a new amqp notification instance
func New(config *model.NotifAmqp, app model.App) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
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

	connString := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, c.cfg.Host, c.cfg.Port)

	conn, err := amqp.Dial(connString)
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
		c.cfg.Queue, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return err
	}

	body, err := buildBody(entry, c.app)
	if err != nil {
		return err
	}

	return ch.Publish(
		c.cfg.Exchange, // exchange
		q.Name,         // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func buildBody(entry model.NotifEntry, app model.App) ([]byte, error) {
	return json.Marshal(struct {
		Version  string        `json:"diun_version"`
		Status   string        `json:"status"`
		Provider string        `json:"provider"`
		Image    string        `json:"image"`
		MIMEType string        `json:"mime_type"`
		Digest   digest.Digest `json:"digest"`
		Created  *time.Time    `json:"created"`
		Platform string        `json:"platform"`
	}{
		Version:  app.Version,
		Status:   string(entry.Status),
		Provider: entry.Provider,
		Image:    entry.Image.String(),
		MIMEType: entry.Manifest.MIMEType,
		Digest:   entry.Manifest.Digest,
		Created:  entry.Manifest.Created,
		Platform: entry.Manifest.Platform,
	})
}
