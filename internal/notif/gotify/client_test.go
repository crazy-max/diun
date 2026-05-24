package gotify

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

func TestSendPostsMessage(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotToken string
	var gotUserAgent string
	var gotContentType string
	var gotPayload struct {
		Message  string                       `json:"message"`
		Title    string                       `json:"title"`
		Priority int                          `json:"priority"`
		Extras   map[string]map[string]string `json:"extras"`
	}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Gotify-Key")
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL + "/api").Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/api/message", gotPath)
	assert.Equal(t, "gotify-token", gotToken)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "docker.io/library/alpine:latest", gotPayload.Title)
	assert.Equal(t, "file update", gotPayload.Message)
	assert.Equal(t, 3, gotPayload.Priority)
	assert.Equal(t, map[string]map[string]string{
		"client::display": {
			"contentType": "text/markdown",
		},
	}, gotPayload.Extras)
}

func TestSendReturnsGotifyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"error":            "bad request",
			"errorCode":        400,
			"errorDescription": "invalid application token",
		}))
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.EqualError(t, err, "400 bad request: invalid application token")
}

func newTestClient(endpoint string) *Client {
	return &Client{
		cfg: &model.NotifGotify{
			Endpoint:      endpoint,
			Token:         "gotify-token",
			Priority:      3,
			Timeout:       new(2 * time.Second),
			TemplateTitle: "{{ .Entry.Image }}",
			TemplateBody:  "{{ .Entry.Provider }} {{ .Entry.Status }}",
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
