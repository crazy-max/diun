package teams

import (
	"encoding/json"
	"fmt"
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

func TestSendPostsMessageCard(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotContentType string
	var gotPayload struct {
		Type       string     `json:"@type"`
		Context    string     `json:"@context"`
		ThemeColor string     `json:"themeColor"`
		Summary    string     `json:"summary"`
		Sections   []Sections `json:"sections"`
	}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "MessageCard", gotPayload.Type)
	assert.Equal(t, "https://schema.org/extensions", gotPayload.Context)
	assert.Equal(t, "0076D7", gotPayload.ThemeColor)
	assert.Equal(t, "file update", gotPayload.Summary)
	require.Len(t, gotPayload.Sections, 1)

	section := gotPayload.Sections[0]
	assert.Equal(t, "file update", section.ActivityTitle)
	assert.Equal(t, fmt.Sprintf("CrazyMax © %d Diun 4.0.0", time.Now().Year()), section.ActivitySubtitle)
	assert.Equal(t, []Fact{
		{Name: "Hostname", Value: "node-1"},
		{Name: "Provider", Value: "file"},
		{Name: "Created", Value: "May 24, 2026 12:34:56 UTC"},
		{Name: "Digest", Value: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{Name: "Platform", Value: "linux/amd64"},
	}, section.Facts)
}

func newTestClient(webhookURL string) *Client {
	renderFacts := true
	timeout := 2 * time.Second
	return &Client{
		cfg: &model.NotifTeams{
			WebhookURL:   webhookURL,
			RenderFacts:  &renderFacts,
			Timeout:      &timeout,
			TemplateBody: "{{ .Entry.Provider }} {{ .Entry.Status }}",
		},
		meta: model.Meta{
			Name:      "Diun",
			Author:    "CrazyMax",
			Version:   "4.0.0",
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

	created := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)
	return model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "file",
		Image:    image,
		Manifest: registry.Manifest{
			Digest:   digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
			Created:  &created,
			Platform: "linux/amd64",
		},
	}
}
