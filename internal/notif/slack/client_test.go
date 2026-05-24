package slack

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

func TestSendPostsWebhookAttachment(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotPayload struct {
		Attachments []struct {
			Color         string `json:"color"`
			AuthorName    string `json:"author_name"`
			AuthorSubname string `json:"author_subname"`
			AuthorLink    string `json:"author_link"`
			AuthorIcon    string `json:"author_icon"`
			Text          string `json:"text"`
			Footer        string `json:"footer"`
			Fields        []struct {
				Title string `json:"title"`
				Value string `json:"value"`
				Short bool   `json:"short"`
			} `json:"fields"`
			Ts json.Number `json:"ts"`
		} `json:"attachments"`
	}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json", gotContentType)
	require.Len(t, gotPayload.Attachments, 1)

	attachment := gotPayload.Attachments[0]
	assert.Equal(t, "#0054ca", attachment.Color)
	assert.Equal(t, "Diun", attachment.AuthorName)
	assert.Equal(t, "github.com/crazy-max/diun", attachment.AuthorSubname)
	assert.Equal(t, "https://example.com/diun", attachment.AuthorLink)
	assert.Equal(t, "https://example.com/logo.png", attachment.AuthorIcon)
	assert.Equal(t, "file update", attachment.Text)
	assert.Equal(t, fmt.Sprintf("CrazyMax © %d Diun 4.0.0", time.Now().Year()), attachment.Footer)
	assert.NotEmpty(t, attachment.Ts)
	assert.Equal(t, []struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short"`
	}{
		{Title: "Hostname", Value: "node-1"},
		{Title: "Provider", Value: "file"},
		{Title: "Created", Value: "May 24, 2026 12:34:56 UTC"},
		{Title: "Digest", Value: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{Title: "Platform", Value: "linux/amd64"},
		{Title: "HubLink", Value: "https://hub.docker.com/r/library/alpine"},
	}, attachment.Fields)
}

func newTestClient(webhookURL string) *Client {
	renderFields := true
	return &Client{
		cfg: &model.NotifSlack{
			WebhookURL:   webhookURL,
			RenderFields: &renderFields,
			TemplateBody: "{{ .Entry.Provider }} {{ .Entry.Status }}",
		},
		meta: model.Meta{
			Name:     "Diun",
			URL:      "https://example.com/diun",
			Logo:     "https://example.com/logo.png",
			Author:   "CrazyMax",
			Version:  "4.0.0",
			Hostname: "node-1",
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
