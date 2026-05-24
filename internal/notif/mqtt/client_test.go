package mqtt

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPublishesRenderedJSON(t *testing.T) {
	mqttClient := &fakeMQTTClient{}
	client := newTestClient(mqttClient)

	err := client.Send(testEntry(t))
	require.NoError(t, err)

	assert.Equal(t, 1, mqttClient.connectCalls)
	assert.Equal(t, "diun/images", mqttClient.topic)
	assert.Equal(t, byte(1), mqttClient.qos)
	assert.False(t, mqttClient.retained)

	var payload struct {
		Status   string `json:"status"`
		Provider string `json:"provider"`
		Image    string `json:"image"`
	}
	require.NoError(t, json.Unmarshal(mqttClient.payload, &payload))
	assert.Equal(t, "update", payload.Status)
	assert.Equal(t, "file", payload.Provider)
	assert.Equal(t, "docker.io/library/alpine:latest", payload.Image)
}

func TestSendReturnsConnectError(t *testing.T) {
	mqttClient := &fakeMQTTClient{
		connectErr: errors.New("connect failed"),
	}
	client := newTestClient(mqttClient)

	err := client.Send(testEntry(t))

	require.EqualError(t, err, "connect failed")
	assert.Equal(t, 1, mqttClient.connectCalls)
	assert.Zero(t, mqttClient.publishCalls)
}

func TestSendReturnsPublishError(t *testing.T) {
	mqttClient := &fakeMQTTClient{
		connected:  true,
		publishErr: errors.New("publish failed"),
	}
	client := newTestClient(mqttClient)

	err := client.Send(testEntry(t))

	require.EqualError(t, err, "publish failed")
	assert.Zero(t, mqttClient.connectCalls)
	assert.Equal(t, 1, mqttClient.publishCalls)
}

type fakeMQTTClient struct {
	connected    bool
	connectCalls int
	publishCalls int
	connectErr   error
	publishErr   error
	topic        string
	qos          byte
	retained     bool
	payload      []byte
}

func (c *fakeMQTTClient) IsConnected() bool {
	return c.connected
}

func (c *fakeMQTTClient) IsConnectionOpen() bool {
	return c.connected
}

func (c *fakeMQTTClient) Connect() MQTT.Token {
	c.connectCalls++
	return fakeToken{err: c.connectErr}
}

func (c *fakeMQTTClient) Disconnect(uint) {
	c.connected = false
}

func (c *fakeMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	c.publishCalls++
	c.topic = topic
	c.qos = qos
	c.retained = retained
	c.payload, _ = payload.([]byte)
	return fakeToken{err: c.publishErr}
}

func (c *fakeMQTTClient) Subscribe(string, byte, MQTT.MessageHandler) MQTT.Token {
	panic("not implemented")
}

func (c *fakeMQTTClient) SubscribeMultiple(map[string]byte, MQTT.MessageHandler) MQTT.Token {
	panic("not implemented")
}

func (c *fakeMQTTClient) Unsubscribe(...string) MQTT.Token {
	panic("not implemented")
}

func (c *fakeMQTTClient) AddRoute(string, MQTT.MessageHandler) {
	panic("not implemented")
}

func (c *fakeMQTTClient) OptionsReader() MQTT.ClientOptionsReader {
	return MQTT.NewOptionsReader(MQTT.NewClientOptions())
}

type fakeToken struct {
	err error
}

func (t fakeToken) Wait() bool {
	return true
}

func (t fakeToken) WaitTimeout(time.Duration) bool {
	return true
}

func (t fakeToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (t fakeToken) Error() error {
	return t.err
}

func newTestClient(mqttClient MQTT.Client) *Client {
	return &Client{
		cfg: &model.NotifMqtt{
			Scheme:   "mqtt",
			Host:     "localhost",
			Port:     1883,
			Username: "mqtt-user",
			Password: "mqtt-password",
			Client:   "diun-test-client",
			Topic:    "diun/images",
			QoS:      1,
		},
		meta: model.Meta{
			Version:  "4.0.0",
			Hostname: "node-1",
		},
		mqttClient: mqttClient,
	}
}

func testEntry(t *testing.T) model.NotifEntry {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "docker.io/library/alpine:latest",
	})
	require.NoError(t, err)

	return model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "file",
		Image:    image,
	}
}
