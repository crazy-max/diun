package mqtt

import (
	"encoding/json"
	"fmt"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active mqtt notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifMqtt
	meta model.Meta
	logger zerolog.Logger
}

// New creates a new mqtt notification instance
func New(config *model.NotifMqtt, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
			logger: log.With().Str("notif", "mqtt").Logger(),
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "mqtt"
}

// Send creates and sends a mqtt notification with an entry
func (c *Client) Send(entry model.NotifEntry) error {
	username, err := utl.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		return err
	}

	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		return err
	}

	broker := fmt.Sprintf("tcp://%s:%d", c.cfg.Host, c.cfg.Port)
	opts := MQTT.NewClientOptions().AddBroker(broker).SetClientID(c.cfg.Client)
	opts.Username = username
	opts.Password = password

	var client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Debug().Msgf("Connected to broker: %s", broker)

	message, err := json.Marshal(struct {
		Version  string     `json:"diun_version"`
		Hostname string     `json:"hostname"`
		Status   string     `json:"status"`
		Provider string     `json:"provider"`
		Image    string     `json:"image"`
		HubLink  string     `json:"hub_link"`
		MIMEType string     `json:"mime_type"`
		Created  *time.Time `json:"created"`
		Platform string     `json:"platform"`
	}{
		Version:  c.meta.Version,
		Hostname: c.meta.Hostname,
		Status:   string(entry.Status),
		Provider: entry.Provider,
		Image:    entry.Image.String(),
		HubLink:  entry.Image.HubLink,
		MIMEType: entry.Manifest.MIMEType,
		Created:  entry.Manifest.Created,
		Platform: entry.Manifest.Platform,
	})
	if err != nil {
		return err
	}

	log.Debug().Msgf("Publishing to topic: %s", c.cfg.Topic)
	token := client.Publish(c.cfg.Topic, byte(c.cfg.QoS), false, message)
	token.Wait()

	return token.Error()
}
