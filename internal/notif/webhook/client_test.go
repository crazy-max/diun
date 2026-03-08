package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/require"
)

func TestSendUsesConfiguredMethod(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotHeader string
	var gotBody []byte
	var gotBodyErr error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotUserAgent = r.Header.Get("User-Agent")
		gotHeader = r.Header.Get("X-Test")
		gotBody, gotBodyErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	image, err := registry.ParseImage(registry.ParseImageOptions{Name: "docker.io/library/alpine:latest"})
	require.NoError(t, err)

	timeout := 2 * time.Second
	c := Client{
		cfg: &model.NotifWebhook{
			Endpoint: ts.URL,
			Method:   http.MethodPut,
			Headers: map[string]string{
				"X-Test": "ok",
			},
			Timeout: &timeout,
		},
		meta: model.Meta{UserAgent: "diun-test"},
	}

	err = c.Send(model.NotifEntry{
		Status:   model.ImageStatusUpdate,
		Provider: "docker",
		Image:    image,
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, gotMethod)
	require.Equal(t, "diun-test", gotUserAgent)
	require.Equal(t, "ok", gotHeader)
	require.NoError(t, gotBodyErr)
	require.NotEmpty(t, gotBody)
	require.True(t, json.Valid(gotBody))
}
