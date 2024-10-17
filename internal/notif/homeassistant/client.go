package mqtt

import (
    "encoding/json"
    "fmt"
    "strings"
    "github.com/crazy-max/diun/v4/internal/model"
    "github.com/crazy-max/diun/v4/internal/notif/notifier"
    "github.com/crazy-max/diun/v4/pkg/utl"
    MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Client represents an active mqtt notification object
type Client struct {
    *notifier.Notifier
    cfg        *model.NotifHomeAssistant
    meta       model.Meta
    mqttClient MQTT.Client
}

// New creates a new mqtt notification instance
func New(config *model.NotifHomeAssistant, meta model.Meta) notifier.Notifier {
    return notifier.Notifier{
        Handler: &Client{
            cfg:  config,
            meta: meta,
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
    if (err != nil) {
        return err
    }

    password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
    if (err != nil) {
        return err
    }

    broker := fmt.Sprintf("%s://%s:%d", c.cfg.Scheme, c.cfg.Host, c.cfg.Port)
    opts := MQTT.NewClientOptions().AddBroker(broker).SetClientID(c.cfg.Client)
    opts.Username = username
    opts.Password = password

    if c.mqttClient == nil {
        c.mqttClient = MQTT.NewClient(opts)
        if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
            return token.Error()
        }
    }

    // Extract the image string
    imageStr := entry.Image.String()
	// Extract the repository name (without version) and sanitize it
	repoName := strings.Split(imageStr, ":")[0]
    sanitizedImage := strings.ReplaceAll(repoName, "/", "_")

    // Define the discovery topic
    discoveryTopic := fmt.Sprintf("%s/%s/%s/%s/config", c.cfg.DiscoveryPrefix, c.cfg.Component, c.cfg.NodeName, sanitizedImage)

    // Create the discovery payload
    discoveryPayload := map[string]interface{}{
        "name":        sanitizedImage,
        "unique_id":   sanitizedImage,
        "state_topic": fmt.Sprintf("%s/%s/%s/%s/config", c.cfg.DiscoveryPrefix, c.cfg.Component, c.cfg.NodeName, sanitizedImage),
        "json_attributes_topic": fmt.Sprintf("%s/%s/%s/%s/config", c.cfg.DiscoveryPrefix, c.cfg.Component, c.cfg.NodeName, sanitizedImage),
        "availability_topic":    "homeassistant/status",
        "device": map[string]interface{}{
            "identifiers":  sanitizedImage,
            "name":         sanitizedImage,
            "sw_version":   "1.0",
            "model":        "MQTT Sensor",
            "manufacturer": "Diun Image Update Notifier",
        },
    }

    payloadBytes, err := json.Marshal(discoveryPayload)
    if err != nil {
        return err
    }

    // Publish the discovery message
    token := c.mqttClient.Publish(discoveryTopic, byte(c.cfg.QoS), true, payloadBytes)
    token.Wait()
    if token.Error() != nil {
        return token.Error()
    }

	// Prepare the state payload
	var statePayload map[string]interface{}
	if entry.Status == "new" {
		statePayload = map[string]interface{}{
			"state":           "New Image",
			"current_version": entry.Image,
			"new_version":     "",
			"icon":            "mdi:package-variant",
		}
	} else if entry.Status == "update" {
		statePayload = map[string]interface{}{
			"state":           "Update Available",
			"current_version": entry.Image,
			"new_version":     entry.Image,
			"icon":            "mdi:package-up",
		}
	} else {
		statePayload = map[string]interface{}{
			"state":           "No Update",
			"current_version": entry.Image,
			"new_version":     "",
			"icon":            "mdi:package-variant-closed",
		}
	}

    statePayloadBytes, err := json.Marshal(statePayload)
    if err != nil {
        return err
    }

    // Publish the state message
    stateTopic := fmt.Sprintf("%s/%s/%s/%s/config", c.cfg.DiscoveryPrefix, c.cfg.Component, c.cfg.NodeName, sanitizedImage)
    token = c.mqttClient.Publish(stateTopic, byte(c.cfg.QoS), false, statePayloadBytes)
    token.Wait()
    if token.Error() != nil {
        return token.Error()
    }

    return nil
}