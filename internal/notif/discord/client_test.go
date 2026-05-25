package discord

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

func TestSendPostsWebhookMessage(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotContentType string
	var gotPayload Message
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "<@123> <@456> file update", gotPayload.Content)
	assert.Equal(t, "Diun", gotPayload.Username)
	assert.Equal(t, "https://example.com/logo.png", gotPayload.AvatarURL)
	require.Len(t, gotPayload.Embeds, 1)

	embed := gotPayload.Embeds[0]
	assert.Equal(t, EmbedAuthor{
		Name:    "Diun",
		URL:     "https://example.com/diun",
		IconURL: "https://example.com/logo.png",
	}, embed.Author)
	assert.Equal(t, fmt.Sprintf("CrazyMax © %d Diun 4.0.0", time.Now().Year()), embed.Footer.Text)
	assert.Equal(t, []EmbedField{
		{Name: "Hostname", Value: "node-1"},
		{Name: "Provider", Value: "file"},
		{Name: "Created", Value: "May 24, 2026 12:34:56 UTC"},
		{Name: "Digest", Value: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{Name: "Platform", Value: "linux/amd64"},
		{Name: "HubLink", Value: "https://hub.docker.com/r/library/alpine"},
	}, embed.Fields)
}

func TestSendRetriesAfterDiscordRateLimit(t *testing.T) {
	var requestCount int
	var gotPayloads []Message
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		var payload Message
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && gotPayloadErr == nil {
			gotPayloadErr = err
		}
		gotPayloads = append(gotPayloads, payload)

		if requestCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited","retry_after":0.001}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)
	assert.Equal(t, 2, requestCount)
	require.Len(t, gotPayloads, 2)
	assert.Equal(t, gotPayloads[0], gotPayloads[1])
}

func TestSendStopsAfterDiscordRateLimitAttempts(t *testing.T) {
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"rate limited","retry_after":0.001}`))
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.ErrorContains(t, err, "unexpected HTTP status 429")
	assert.Equal(t, discordMaxRateLimitAttempts, requestCount)
}

func TestDiscordRetryAfterUsesHeader(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
	}
	resp.Header.Set("Retry-After", "0.25")

	assert.Equal(t, 250*time.Millisecond, discordRetryAfter(resp, nil))
}

func TestDiscordRetryAfterUsesBodyFallback(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
	}

	assert.Equal(t, 500*time.Millisecond, discordRetryAfter(resp, []byte(`{"retry_after":0.5}`)))
}

func TestDiscordRetryAfterUsesRateLimitResetAfterFallback(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
	}
	resp.Header.Set("X-RateLimit-Reset-After", "0.75")

	assert.Equal(t, 750*time.Millisecond, discordRetryAfter(resp, nil))
}

func newTestClient(webhookURL string) *Client {
	return &Client{
		cfg: &model.NotifDiscord{
			WebhookURL:   webhookURL,
			Mentions:     []string{"<@123>", "<@456>"},
			RenderEmbeds: new(true),
			RenderFields: new(true),
			Timeout:      new(2 * time.Second),
			TemplateBody: "{{ .Entry.Provider }} {{ .Entry.Status }}",
		},
		meta: model.Meta{
			Name:      "Diun",
			URL:       "https://example.com/diun",
			Logo:      "https://example.com/logo.png",
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
