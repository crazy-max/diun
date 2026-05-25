package matrix

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRetriesAfterMatrixRateLimit(t *testing.T) {
	var sendCount int
	var gotDecodeErr error
	var gotEncodeErr error
	var gotPayloads []matrixMessagePayload
	var gotSendPaths []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/login"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{
				"access_token": "access-token",
				"device_id":    "DEVICE",
				"user_id":      "@diun:example.com",
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.Contains(r.URL.Path, "/join/"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{
				"room_id": "!alerts:example.com",
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.Contains(r.URL.Path, "/send/m.room.message/"):
			sendCount++
			gotSendPaths = append(gotSendPaths, r.URL.Path)
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

			var payload matrixMessagePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && gotDecodeErr == nil {
				gotDecodeErr = err
			}
			gotPayloads = append(gotPayloads, payload)

			if sendCount == 1 {
				w.Header().Set("Retry-After", "0")
				if err := writeMatrixRateLimitResponse(w); err != nil && gotEncodeErr == nil {
					gotEncodeErr = err
				}
				return
			}
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{
				"event_id": "$event",
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.HasSuffix(r.URL.Path, "/logout"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.NoError(t, err)
	require.NoError(t, gotDecodeErr)
	require.NoError(t, gotEncodeErr)
	assert.Equal(t, 2, sendCount)
	require.Len(t, gotPayloads, 2)
	require.Len(t, gotSendPaths, 2)
	assert.Equal(t, gotPayloads[0], gotPayloads[1])
	assert.Equal(t, gotSendPaths[0], gotSendPaths[1])
	assert.Equal(t, matrixMessagePayload{
		Body:          "file update",
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: "file update",
	}, gotPayloads[0])
}

func TestSendStopsAfterMatrixRateLimitAttempts(t *testing.T) {
	var sendCount int
	var gotEncodeErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/login"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{
				"access_token": "access-token",
				"device_id":    "DEVICE",
				"user_id":      "@diun:example.com",
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.Contains(r.URL.Path, "/join/"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{
				"room_id": "!alerts:example.com",
			}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.Contains(r.URL.Path, "/send/m.room.message/"):
			sendCount++
			w.Header().Set("Retry-After", "0")
			if err := writeMatrixRateLimitResponse(w); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		case strings.HasSuffix(r.URL.Path, "/logout"):
			if err := writeMatrixResponse(w, http.StatusOK, map[string]any{}); err != nil && gotEncodeErr == nil {
				gotEncodeErr = err
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	err := newTestClient(ts.URL).Send(testEntry(t))

	require.ErrorContains(t, err, "failed to submit message to Matrix")
	require.NoError(t, gotEncodeErr)
	assert.Equal(t, matrixMaxRateLimitAttempts, sendCount)
}

type matrixMessagePayload struct {
	Body          string `json:"body"`
	MsgType       string `json:"msgtype"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

func newTestClient(homeserverURL string) *Client {
	return &Client{
		cfg: &model.NotifMatrix{
			HomeserverURL: homeserverURL,
			User:          "@diun:example.com",
			Password:      "password",
			RoomID:        "!alerts:example.com",
			MsgType:       model.NotifMatrixMsgTypeNotice,
			TemplateBody:  "{{ .Entry.Provider }} {{ .Entry.Status }}",
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

func writeMatrixRateLimitResponse(w http.ResponseWriter) error {
	return writeMatrixResponse(w, http.StatusTooManyRequests, map[string]any{
		"errcode": "M_LIMIT_EXCEEDED",
		"error":   "rate limited",
	})
}

func writeMatrixResponse(w http.ResponseWriter, statusCode int, body map[string]any) error {
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(body)
}
