package matrix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestSendAuthentication(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("file-token"), 0o600))
	missingFile := filepath.Join(t.TempDir(), "missing")

	tests := []struct {
		name        string
		credentials model.NotifMatrix
		failPath    string
		wantToken   string
		wantErr     string
		wantLogin   int
		wantLogout  int
	}{
		{
			name: "token without username",
			credentials: model.NotifMatrix{
				Token: "inline-token",
			},
			wantToken: "inline-token",
		},
		{
			name: "token file",
			credentials: model.NotifMatrix{
				TokenFile: tokenFile,
			},
			wantToken: "file-token",
		},
		{
			name: "token takes precedence over file and password credentials",
			credentials: model.NotifMatrix{
				Token:        "inline-token",
				TokenFile:    missingFile,
				UserFile:     missingFile,
				Password:     "unused-password",
				PasswordFile: missingFile,
			},
			wantToken: "inline-token",
		},
		{
			name: "unreadable token file",
			credentials: model.NotifMatrix{
				TokenFile: missingFile,
			},
			wantErr: "cannot retrieve token secret for Matrix notifier",
		},
		{
			name: "token survives join failure",
			credentials: model.NotifMatrix{
				Token: "inline-token",
			},
			failPath:  "/join/",
			wantToken: "inline-token",
			wantErr:   "failed to join room",
		},
		{
			name: "token survives send failure",
			credentials: model.NotifMatrix{
				Token: "inline-token",
			},
			failPath:  "/send/",
			wantToken: "inline-token",
			wantErr:   "failed to submit message to Matrix",
		},
		{
			name: "password login",
			credentials: model.NotifMatrix{
				User:     "@diun:example.com",
				Password: "password",
			},
			wantToken:  "login-token",
			wantLogin:  1,
			wantLogout: 1,
		},
		{
			name: "password session logged out after send failure",
			credentials: model.NotifMatrix{
				User:     "@diun:example.com",
				Password: "password",
			},
			failPath:   "/send/",
			wantToken:  "login-token",
			wantErr:    "failed to submit message to Matrix",
			wantLogin:  1,
			wantLogout: 1,
		},
		{
			name: "failed login does not log out",
			credentials: model.NotifMatrix{
				User:     "@diun:example.com",
				Password: "password",
			},
			failPath:  "/login",
			wantErr:   "failed to authenticate Matrix user",
			wantLogin: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginCount, logoutCount, joinCount, sendCount int
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				var response map[string]any
				switch {
				case strings.HasSuffix(r.URL.Path, "/login"):
					loginCount++
					var login struct {
						Identifier struct {
							User string `json:"user"`
						} `json:"identifier"`
						Password string `json:"password"`
					}
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&login))
					assert.Equal(t, tt.credentials.User, login.Identifier.User)
					assert.Equal(t, tt.credentials.Password, login.Password)
					response = map[string]any{
						"access_token": "login-token",
						"user_id":      "@diun:example.com",
					}
				case strings.HasSuffix(r.URL.Path, "/logout"):
					logoutCount++
					response = map[string]any{}
				case strings.Contains(r.URL.Path, "/join/"):
					joinCount++
					response = map[string]any{"room_id": "!alerts:example.com"}
				case strings.Contains(r.URL.Path, "/send/m.room.message/"):
					sendCount++
					var payload matrixMessagePayload
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
					assert.Equal(t, matrixMessagePayload{
						Body:          "file update",
						MsgType:       "m.notice",
						Format:        "org.matrix.custom.html",
						FormattedBody: "file update",
					}, payload)
					response = map[string]any{"event_id": "$event"}
				default:
					t.Errorf("unexpected Matrix request: %s", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if !strings.HasSuffix(r.URL.Path, "/login") {
					assert.Equal(t, "Bearer "+tt.wantToken, r.Header.Get("Authorization"))
				}
				if tt.failPath != "" && strings.Contains(r.URL.Path, tt.failPath) {
					assert.NoError(t, writeMatrixResponse(w, http.StatusForbidden, map[string]any{
						"errcode": "M_FORBIDDEN",
						"error":   "forbidden",
					}))
					return
				}
				assert.NoError(t, writeMatrixResponse(w, http.StatusOK, response))
			}))
			defer ts.Close()

			c := newTestClient(ts.URL)
			c.cfg.User = tt.credentials.User
			c.cfg.UserFile = tt.credentials.UserFile
			c.cfg.Password = tt.credentials.Password
			c.cfg.PasswordFile = tt.credentials.PasswordFile
			c.cfg.Token = tt.credentials.Token
			c.cfg.TokenFile = tt.credentials.TokenFile
			err := c.Send(testEntry(t))
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 1, joinCount)
				assert.Equal(t, 1, sendCount)
			}
			assert.Equal(t, tt.wantLogin, loginCount)
			assert.Equal(t, tt.wantLogout, logoutCount)
			if tt.wantToken == "" {
				assert.Zero(t, joinCount)
				assert.Zero(t, sendCount)
			}
		})
	}
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
