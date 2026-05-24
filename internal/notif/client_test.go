package notif

import (
	"errors"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/notif/notifier"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithoutConfigReturnsEmptyClient(t *testing.T) {
	client, err := New(nil, model.Meta{})
	require.NoError(t, err)

	assert.Empty(t, client.List())
}

func TestNewBuildsConfiguredNotifiers(t *testing.T) {
	client, err := New(&model.Notif{
		Ntfy:    &model.NotifNtfy{},
		Webhook: &model.NotifWebhook{},
	}, model.Meta{})
	require.NoError(t, err)

	require.Len(t, client.List(), 2)
	assert.Equal(t, "ntfy", client.List()[0].Name())
	assert.Equal(t, "webhook", client.List()[1].Name())
}

func TestSendDispatchesAllNotifiers(t *testing.T) {
	entry := model.NotifEntry{
		Status: model.ImageStatusUpdate,
		Image:  parseTestImage(t, "crazymax/diun:1.2.3"),
	}
	first := &fakeNotifier{name: "first"}
	failing := &fakeNotifier{name: "failing", err: errors.New("boom")}
	last := &fakeNotifier{name: "last"}
	client := &Client{
		notifiers: []notifier.Notifier{
			{Handler: first},
			{Handler: failing},
			{Handler: last},
		},
	}

	client.Send(entry)

	assert.Equal(t, []model.NotifEntry{entry}, first.entries)
	assert.Equal(t, []model.NotifEntry{entry}, failing.entries)
	assert.Equal(t, []model.NotifEntry{entry}, last.entries)
}

type fakeNotifier struct {
	name    string
	err     error
	entries []model.NotifEntry
}

func (n *fakeNotifier) Name() string {
	return n.name
}

func (n *fakeNotifier) Send(entry model.NotifEntry) error {
	n.entries = append(n.entries, entry)
	return n.err
}

func parseTestImage(t *testing.T, name string) registry.Image {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{Name: name})
	require.NoError(t, err)
	return image
}
