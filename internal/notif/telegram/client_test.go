package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRetriesAfterTelegramRateLimit(t *testing.T) {
	var sendCount int
	var gotEncodeErr error
	var gotParseErr error
	var gotTexts []string
	var gotThreadIDs []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			if err := writeTelegramResponse(w, http.StatusOK, map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":         123,
					"is_bot":     true,
					"first_name": "Diun",
				},
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.HasSuffix(r.URL.Path, "/sendMessage"):
			sendCount++
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if err := r.ParseMultipartForm(32 << 20); err != nil && gotParseErr == nil {
				gotParseErr = err
			}
			gotTexts = append(gotTexts, r.FormValue("text"))
			gotThreadIDs = append(gotThreadIDs, r.FormValue("message_thread_id"))

			if sendCount == 1 {
				if err := writeTelegramRateLimitResponse(w); err != nil && gotEncodeErr == nil {
					gotEncodeErr = err
				}
				return
			}
			if err := writeTelegramResponse(w, http.StatusOK, map[string]any{
				"ok":     true,
				"result": map[string]any{},
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, err)
	require.NoError(t, gotEncodeErr)
	require.NoError(t, gotParseErr)
	assert.Equal(t, 2, sendCount)
	assert.Equal(t, []string{"file update", "file update"}, gotTexts)
	assert.Equal(t, []string{"456", "456"}, gotThreadIDs)
}

func TestSendStopsAfterTelegramRateLimitAttempts(t *testing.T) {
	var sendCount int
	var gotEncodeErr error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			if err := writeTelegramResponse(w, http.StatusOK, map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":         123,
					"is_bot":     true,
					"first_name": "Diun",
				},
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.HasSuffix(r.URL.Path, "/sendMessage"):
			sendCount++
			if err := writeTelegramRateLimitResponse(w); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.ErrorContains(t, err, "unable to sendMessage")
	require.NoError(t, gotEncodeErr)
	assert.Equal(t, telegramMaxRateLimitAttempts, sendCount)
}

func TestParseChatIDs(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string
		expected []chatID
		err      error
	}{
		{
			name:    "valid chat IDS",
			entries: []string{"8547439", "1234567"},
			expected: []chatID{
				{id: 8547439},
				{id: 1234567},
			},
			err: nil,
		},
		{
			name:    "valid strings with topics",
			entries: []string{"567891234:25", "891256734:25;12"},
			expected: []chatID{
				{id: 567891234, topics: []int64{25}},
				{id: 891256734, topics: []int64{25, 12}},
			},
			err: nil,
		},
		{
			name:     "invalid format",
			entries:  []string{"invalid_format"},
			expected: nil,
			err:      errors.New(`invalid chat ID: strconv.ParseInt: parsing "invalid_format": invalid syntax`),
		},
		{
			name:     "empty string",
			entries:  []string{""},
			expected: nil,
			err:      errors.New(`invalid chat ID: strconv.ParseInt: parsing "": invalid syntax`),
		},
		{
			name:     "string with invalid topic",
			entries:  []string{"567891234:invalid"},
			expected: nil,
			err:      errors.New(`invalid topic "invalid" for chat ID 567891234: strconv.ParseInt: parsing "invalid": invalid syntax`),
		},
		{
			name:     "invalid format with too many parts",
			entries:  []string{"567891234:25:extra"},
			expected: nil,
			err:      errors.New(`invalid chat ID "567891234:25:extra"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseChatIDs(tt.entries)
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, res)
		})
	}
}

func newTestClient(apiURL string) *Client {
	return &Client{
		cfg: &model.NotifTelegram{
			APIURL:              apiURL,
			Token:               "123:abc",
			ChatIDs:             []string{"123:456"},
			TemplateBody:        "{{ .Entry.Provider }} {{ .Entry.Status }}",
			DisableNotification: new(true),
		},
		meta: model.Meta{
			Name:     "Diun",
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

func writeTelegramRateLimitResponse(w http.ResponseWriter) error {
	return writeTelegramResponse(w, http.StatusTooManyRequests, map[string]any{
		"ok":          false,
		"error_code":  http.StatusTooManyRequests,
		"description": "Too Many Requests: retry later",
		"parameters": map[string]any{
			"retry_after": 0,
		},
	})
}

func writeTelegramResponse(w http.ResponseWriter, statusCode int, body map[string]any) error {
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(body)
}
