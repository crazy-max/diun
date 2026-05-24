package elasticsearch

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

func TestSendIndexesNotificationDocument(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotUsername string
	var gotPassword string
	var gotAuthOK bool
	var gotUserAgent string
	var gotContentType string
	var gotPayload map[string]any
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotUsername, gotPassword, gotAuthOK = r.BasicAuth()
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL + "/logs").Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/logs/diun-events/_doc", gotPath)
	assert.True(t, gotAuthOK)
	assert.Equal(t, "elastic-user", gotUsername)
	assert.Equal(t, "elastic-password", gotPassword)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "diun-test-client", gotPayload["client"])
	assert.Equal(t, "update", gotPayload["status"])
	assert.Equal(t, "file", gotPayload["provider"])
	assert.Equal(t, "docker.io/library/alpine:latest", gotPayload["image"])
	timestamp, ok := gotPayload["@timestamp"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, timestamp)
	_, err = time.Parse(time.RFC3339Nano, timestamp)
	require.NoError(t, err)
}

func TestSendReturnsElasticsearchError(t *testing.T) {
	var encodeErr error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		encodeErr = json.NewEncoder(w).Encode(map[string]any{
			"status": 409,
			"error": map[string]any{
				"type":   "version_conflict_engine_exception",
				"reason": "document already exists",
			},
		})
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, encodeErr)
	require.EqualError(t, err, "409 version_conflict_engine_exception: document already exists")
}

func newTestClient(address string) *Client {
	return &Client{
		cfg: &model.NotifElasticsearch{
			Address:  address,
			Username: "elastic-user",
			Password: "elastic-password",
			Client:   "diun-test-client",
			Index:    "diun-events",
			Timeout:  new(2 * time.Second),
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
