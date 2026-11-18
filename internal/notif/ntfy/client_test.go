package ntfy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/require"
)

func TestSendHeaders(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		token       string
		wantAuth    string
		wantTokenID string
		wantToken   string
	}{
		{
			name: "no custom headers",
		},
		{
			name: "proxy headers and ntfy token",
			headers: map[string]string{
				"p-access-token-id": "proxy-id",
				"p-access-token":    "proxy-token",
				"authorization":     "custom-auth",
				"content-type":      "text/plain",
				"user-agent":        "custom-agent",
			},
			token:       "ntfy-token",
			wantAuth:    "Bearer ntfy-token",
			wantTokenID: "proxy-id",
			wantToken:   "proxy-token",
		},
		{
			name: "custom authorization without ntfy token",
			headers: map[string]string{
				"authorization": "custom-auth",
			},
			wantAuth: "custom-auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHeaders http.Header
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeaders = r.Header.Clone()
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			cfg := (&model.NotifNtfy{}).GetDefaults()
			cfg.Endpoint = ts.URL
			cfg.Topic = "diun"
			cfg.Headers = tt.headers
			cfg.Token = tt.token
			image, err := registry.ParseImage(registry.ParseImageOptions{Name: "alpine:latest"})
			require.NoError(t, err)

			err = New(cfg, model.Meta{UserAgent: "diun-test"}).Send(model.NotifEntry{
				Status:   model.ImageStatusUpdate,
				Provider: "docker",
				Image:    image,
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantTokenID, gotHeaders.Get("P-Access-Token-Id"))
			require.Equal(t, tt.wantToken, gotHeaders.Get("P-Access-Token"))
			require.Equal(t, tt.wantAuth, gotHeaders.Get("Authorization"))
			require.Equal(t, "application/json", gotHeaders.Get("Content-Type"))
			require.Equal(t, "diun-test", gotHeaders.Get("User-Agent"))
			if tt.wantAuth != "" {
				require.Equal(t, []string{tt.wantAuth}, gotHeaders.Values("Authorization"))
			}
		})
	}
}

func TestSendSetsIcon(t *testing.T) {
	tests := []struct {
		name     string
		icon     string
		metaLogo string
		wantIcon string
	}{
		{
			name:     "configured icon",
			icon:     "https://example.com/custom.png",
			metaLogo: "https://example.com/logo.png",
			wantIcon: "https://example.com/custom.png",
		},
		{
			name:     "meta logo fallback",
			metaLogo: "https://example.com/logo.png",
			wantIcon: "https://example.com/logo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotUserAgent string
			var gotPayload struct {
				Topic    string   `json:"topic"`
				Priority int      `json:"priority"`
				Tags     []string `json:"tags"`
				Icon     string   `json:"icon"`
				Click    string   `json:"click"`
				Markdown bool     `json:"markdown"`
			}
			var gotPayloadErr error

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotUserAgent = r.Header.Get("User-Agent")
				gotPayloadErr = json.NewDecoder(r.Body).Decode(&gotPayload)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			image, err := registry.ParseImage(registry.ParseImageOptions{Name: "docker.io/library/alpine:latest"})
			require.NoError(t, err)

			c := Client{
				cfg: &model.NotifNtfy{
					Endpoint:      ts.URL,
					Topic:         "diun",
					Priority:      3,
					Tags:          []string{"package"},
					Icon:          tt.icon,
					Click:         "https://dock.example.com/compose/{{ .Entry.Metadata.stack }}",
					Timeout:       new(2 * time.Second),
					TemplateTitle: "{{ .Entry.Image }}",
					TemplateBody:  "body",
				},
				meta: model.Meta{
					Logo:      tt.metaLogo,
					UserAgent: "diun-test",
				},
			}

			err = c.Send(model.NotifEntry{
				Status:   model.ImageStatusUpdate,
				Provider: "docker",
				Image:    image,
				Metadata: map[string]string{
					"stack": "alpine",
				},
			})
			require.NoError(t, err)
			require.NoError(t, gotPayloadErr)
			require.Equal(t, http.MethodPost, gotMethod)
			require.Equal(t, "diun-test", gotUserAgent)
			require.Equal(t, "diun", gotPayload.Topic)
			require.Equal(t, 3, gotPayload.Priority)
			require.Equal(t, []string{"package"}, gotPayload.Tags)
			require.Equal(t, tt.wantIcon, gotPayload.Icon)
			require.Equal(t, "https://dock.example.com/compose/alpine", gotPayload.Click)
			require.True(t, gotPayload.Markdown)
		})
	}
}
