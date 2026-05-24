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

			timeout := 2 * time.Second
			c := Client{
				cfg: &model.NotifNtfy{
					Endpoint:      ts.URL,
					Topic:         "diun",
					Priority:      3,
					Tags:          []string{"package"},
					Icon:          tt.icon,
					Timeout:       &timeout,
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
			})
			require.NoError(t, err)
			require.NoError(t, gotPayloadErr)
			require.Equal(t, http.MethodPost, gotMethod)
			require.Equal(t, "diun-test", gotUserAgent)
			require.Equal(t, "diun", gotPayload.Topic)
			require.Equal(t, 3, gotPayload.Priority)
			require.Equal(t, []string{"package"}, gotPayload.Tags)
			require.Equal(t, tt.wantIcon, gotPayload.Icon)
			require.True(t, gotPayload.Markdown)
		})
	}
}
