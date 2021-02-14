package mqtt

import (
	"fmt"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/msg"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/utl"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents an active mqtt notification object
type Client struct {
	*notifier.Notifier
	cfg        *model.NotifMqtt
	meta       model.Meta
	logger     zerolog.Logger
	mqttClient MQTT.Client
}

// New creates a new mqtt notification instance
func New(config *model.NotifMqtt, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:    config,
			meta:   meta,
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

	if c.mqttClient == nil {
		c.mqttClient = MQTT.NewClient(opts)
	}
	if !c.mqttClient.IsConnected() {
		if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	message, err := msg.New(msg.Options{
		Meta:  c.meta,
		Entry: entry,
	})
	if err != nil {
		return err
	}

	body, err := message.RenderJSON()
	if err != nil {
		return err
	}

	token := c.mqttClient.Publish(c.cfg.Topic, byte(c.cfg.QoS), false, body)
	token.Wait()
	return token.Error()
}
