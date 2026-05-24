package rocketchat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPostsChatMessage(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotToken string
	var gotUserID string
	var gotUserAgent string
	var gotContentType string
	var gotPayload Message
	var gotPayloadErr error
	var encodeErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Auth-Token")
		gotUserID = r.Header.Get("X-User-Id")
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		encodeErr = json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer ts.Close()

	err := newTestClient(ts.URL + "/root").Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)
	require.NoError(t, encodeErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/root/api/v1/chat.postMessage", gotPath)
	assert.Equal(t, "rocket-token", gotToken)
	assert.Equal(t, "rocket-user", gotUserID)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "Diun", gotPayload.Alias)
	assert.Equal(t, "https://example.com/logo.png", gotPayload.Avatar)
	assert.Equal(t, "#ops", gotPayload.Channel)
	assert.Equal(t, "docker.io/library/alpine:latest", gotPayload.Text)
	require.Len(t, gotPayload.Attachments, 1)

	attachment := gotPayload.Attachments[0]
	assert.Equal(t, "file update", attachment.Text)
	assert.NotEmpty(t, attachment.Ts)
	assert.Equal(t, []AttachmentField{
		{Title: "Hostname", Value: "node-1"},
		{Title: "Provider", Value: "file"},
		{Title: "Created", Value: "May 24, 2026 12:34:56 UTC"},
		{Title: "Digest", Value: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{Title: "Platform", Value: "linux/amd64"},
		{Title: "HubLink", Value: "https://hub.docker.com/r/library/alpine"},
	}, attachment.Fields)
}

func TestSendReturnsRocketChatError(t *testing.T) {
	var encodeErr error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		encodeErr = json.NewEncoder(w).Encode(map[string]any{
			"success":   false,
			"error":     "invalid room",
			"errorType": "error-invalid-room",
		})
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, encodeErr)
	require.EqualError(t, err, "unexpected HTTP error 400: error-invalid-room")
}

func newTestClient(endpoint string) *Client {
	return &Client{
		cfg: &model.NotifRocketChat{
			Endpoint:         endpoint,
			Channel:          "#ops",
			UserID:           "rocket-user",
			Token:            "rocket-token",
			RenderAttachment: new(true),
			Timeout:          new(2 * time.Second),
			TemplateTitle:    "{{ .Entry.Image }}",
			TemplateBody:     "{{ .Entry.Provider }} {{ .Entry.Status }}",
		},
		meta: model.Meta{
			Name:      "Diun",
			Logo:      "https://example.com/logo.png",
			UserAgent: "diun-test",
			Hostname:  "node-1",
		},
	}
}

func testEntry(t *testing.T) model.NotifEntry {
	t.Helper()

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "docker.io/library/alpine:latest",
	})
	require.NoError(t, err)
	image.HubLink = "https://hub.docker.com/r/library/alpine"

	return model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "file",
		Image:    image,
		Manifest: registry.Manifest{
			Digest:   digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
			Created:  new(time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)),
			Platform: "linux/amd64",
		},
	}
}
