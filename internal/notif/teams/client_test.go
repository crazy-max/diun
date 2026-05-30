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
	var gotPayload messageCardPayload
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

func TestSendPostsAdaptiveCard(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotContentType string
	var gotPayload adaptiveCardPayload
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClientWithCardType(ts.URL, model.NotifTeamsCardTypeAdaptiveCard).Send(testEntry(t))
	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "diun-test", gotUserAgent)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "message", gotPayload.Type)
	require.Len(t, gotPayload.Attachments, 1)

	attachment := gotPayload.Attachments[0]
	assert.Equal(t, "application/vnd.microsoft.card.adaptive", attachment.ContentType)
	assert.Equal(t, "http://adaptivecards.io/schemas/adaptive-card.json", attachment.Content.Schema)
	assert.Equal(t, "AdaptiveCard", attachment.Content.Type)
	assert.Equal(t, "1.2", attachment.Content.Version)
	require.Len(t, attachment.Content.Body, 3)

	assert.Equal(t, adaptiveCardElement{
		Type:   "TextBlock",
		Text:   "file update",
		Wrap:   true,
		Weight: "Bolder",
		Color:  "Accent",
	}, attachment.Content.Body[0])
	assert.Equal(t, adaptiveCardElement{
		Type: "FactSet",
		Facts: []adaptiveCardFact{
			{Title: "Hostname", Value: "node-1"},
			{Title: "Provider", Value: "file"},
			{Title: "Created", Value: "May 24, 2026 12:34:56 UTC"},
			{Title: "Digest", Value: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			{Title: "Platform", Value: "linux/amd64"},
		},
	}, attachment.Content.Body[1])
	assert.Equal(t, adaptiveCardElement{
		Type:     "TextBlock",
		Text:     fmt.Sprintf("CrazyMax © %d Diun 4.0.0", time.Now().Year()),
		Wrap:     true,
		Size:     "Small",
		IsSubtle: true,
		Spacing:  "Small",
	}, attachment.Content.Body[2])
}

func TestSendReturnsTeamsErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to deliver message"))
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.ErrorContains(t, err, "unexpected HTTP status 500: failed to deliver message")
}

func TestSendRetriesAfterTeamsRateLimit(t *testing.T) {
	var requestCount int
	var gotPayloads []map[string]interface{}
	var gotPayloadErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		var payload map[string]interface{}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&payload); err != nil && gotPayloadErr == nil {
			gotPayloadErr = err
		}
		gotPayloads = append(gotPayloads, payload)

		if requestCount == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(teamsRateLimitMessage))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, err)
	require.NoError(t, gotPayloadErr)
	assert.Equal(t, 2, requestCount)
	require.Len(t, gotPayloads, 2)
	assert.Equal(t, gotPayloads[0], gotPayloads[1])
}

func TestSendStopsAfterTeamsRateLimitAttempts(t *testing.T) {
	var requestCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(teamsRateLimitMessage))
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.ErrorContains(t, err, "unexpected Teams response: Microsoft Teams endpoint returned HTTP error 429")
	assert.Equal(t, teamsMaxRateLimitAttempts, requestCount)
}

func newTestClient(webhookURL string) *Client {
	return newTestClientWithCardType(webhookURL, model.NotifTeamsCardTypeMessageCard)
}

func newTestClientWithCardType(webhookURL string, cardType model.NotifTeamsCardType) *Client {
	return &Client{
		cfg: &model.NotifTeams{
			WebhookURL:   webhookURL,
			CardType:     cardType,
			RenderFacts:  new(true),
			Timeout:      new(2 * time.Second),
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
