package signalrest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPostsSignalMessage(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotContentType string
	var gotHeader string
	var gotPayload struct {
		Message    string   `json:"message"`
		Number     string   `json:"number"`
		Recipients []string `json:"recipients"`
		TextMode   string   `json:"text_mode"`
	}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotHeader = r.Header.Get("X-Test")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "ok", gotHeader)
	assert.Equal(t, "file update", gotPayload.Message)
	assert.Equal(t, "+15551234567", gotPayload.Number)
	assert.Equal(t, []string{"+15550000001", "+15550000002"}, gotPayload.Recipients)
	assert.Equal(t, "styled", gotPayload.TextMode)
}

func newTestClient(endpoint string) *Client {
	timeout := 2 * time.Second
	return &Client{
		cfg: &model.NotifSignalRest{
			Endpoint:   endpoint,
			Number:     "+15551234567",
			Recipients: []string{"+15550000001", "+15550000002"},
			Headers: map[string]string{
				"X-Test": "ok",
			},
			Timeout:      &timeout,
			TemplateBody: "{{ .Entry.Provider }} {{ .Entry.Status }}",
			TextMode:     "styled",
		},
		meta: model.Meta{
			UserAgent: "diun-test",
		},
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
