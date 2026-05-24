package apprise

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

func TestSendPostsNotifyPayload(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotRawQuery string
	var gotUserAgent string
	var gotContentType string
	var gotPayload struct {
		Body  string   `json:"body"`
		Title string   `json:"title"`
		Tags  []string `json:"tags"`
		URLs  []string `json:"urls"`
	}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL + "/api?format=json").Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/api/notify/apprise-token", gotPath)
	assert.Equal(t, "format=json", gotRawQuery)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "file update", gotPayload.Body)
	assert.Equal(t, "docker.io/library/alpine:latest", gotPayload.Title)
	assert.Equal(t, []string{"ops", "registry"}, gotPayload.Tags)
	assert.Equal(t, []string{"mailto://ops@example.com"}, gotPayload.URLs)
}

func TestSendReturnsAppriseError(t *testing.T) {
	var encodeErr error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		encodeErr = json.NewEncoder(w).Encode(map[string]any{
			"error":            "bad request",
			"errorCode":        400,
			"errorDescription": "invalid notification URL",
		})
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, encodeErr)
	require.EqualError(t, err, "400 bad request: invalid notification URL")
}

func newTestClient(endpoint string) *Client {
	return &Client{
		cfg: &model.NotifApprise{ //nolint:gosec // fixture token is test data.
			Endpoint:      endpoint,
			Token:         "apprise-token",
			Tags:          []string{"ops", "registry"},
			URLs:          []string{"mailto://ops@example.com"},
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
