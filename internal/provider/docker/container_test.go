package docker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContainerPreservesConfiguredTag(t *testing.T) {
	for _, tc := range []struct {
		name       string
		listedID   bool
		tagMissing bool
	}{
		{
			name: "unchanged tag",
		},
		{
			name:     "tag moved after pull",
			listedID: true,
		},
		{
			name:       "tag removed",
			listedID:   true,
			tagMissing: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			imageID := "sha256:" + strings.Repeat("a", 64)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case strings.HasSuffix(r.URL.Path, "/version"):
					_, _ = w.Write([]byte(`{"ApiVersion":"1.52"}`))
				case strings.HasSuffix(r.URL.Path, "/containers/json"):
					imageName := "alpine:latest"
					if tc.listedID {
						imageName = imageID
					}
					_ = json.NewEncoder(w).Encode([]container.Summary{{ID: "watched", Image: imageName, ImageID: imageID, Labels: map[string]string{"diun.enable": "true"}}})
				case strings.HasSuffix(r.URL.Path, "/containers/watched/json"):
					_, _ = w.Write([]byte(`{"Config":{"Image":"alpine:latest"}}`))
				case strings.HasSuffix(r.URL.Path, "/images/alpine:latest/json"):
					if tc.tagMissing {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message":"tag missing"}`))
						return
					}
					_, _ = w.Write([]byte(`{"RepoDigests":["alpine@` + imageID + `"]}`))
				case strings.HasSuffix(r.URL.Path, "/images/"+imageID+"/json"):
					if tc.tagMissing {
						_, _ = w.Write([]byte(`{"RepoDigests":["alpine@` + imageID + `"]}`))
					} else {
						_, _ = w.Write([]byte(`{"RepoDigests":[]}`))
					}
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			cfg := (&model.PrdDocker{}).GetDefaults()
			cfg.Endpoint = server.URL
			cfg.APIVersion = "1.52"
			jobs := New(cfg, nil).ListJob()
			require.Len(t, jobs, 1)
			require.Equal(t, "alpine:latest", jobs[0].Image.Name)
		})
	}
}

func TestMetadataFormatsContainerSummary(t *testing.T) {
	created := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC).Unix()

	got := metadata(container.Summary{
		ID:         "container-id",
		Names:      []string{"/web", "/worker"},
		Image:      "nginx:1.25.0",
		Command:    "nginx -g daemon off;",
		Created:    created,
		State:      "running",
		Status:     "Up 5 minutes",
		SizeRw:     1024,
		SizeRootFs: 2048,
	})

	assert.Equal(t, map[string]string{
		"ctn_id":        "container-id",
		"ctn_names":     "web,worker",
		"ctn_image":     "nginx:1.25.0",
		"ctn_image_tag": "1.25.0",
		"ctn_command":   "nginx -g daemon off;",
		"ctn_createdat": time.Unix(created, 0).String(),
		"ctn_state":     "running",
		"ctn_status":    "Up 5 minutes",
		"ctn_size":      "1.02kB (virtual 2.05kB)",
	}, got)
}

func TestMetadataImageTagParsing(t *testing.T) {
	testCases := []struct {
		name        string
		image       string
		expectedTag string
	}{
		{
			name:        "tagged image",
			image:       "crazymax/diun:4.29.0",
			expectedTag: "4.29.0",
		},
		{
			name:        "untagged defaults to latest",
			image:       "crazymax/diun",
			expectedTag: "latest",
		},
		{
			name:        "custom registry with port and tag",
			image:       "registry.example.com:5000/app:v1.0.0-alpine",
			expectedTag: "v1.0.0-alpine",
		},
		{
			name:        "empty image name",
			image:       "",
			expectedTag: "",
		},
		{
			name:        "image ID has no tag",
			image:       "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expectedTag: "",
		},
		{
			name:        "bare image ID has no tag",
			image:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expectedTag: "",
		},
		{
			name:        "digest reference has no tag",
			image:       "nginx@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expectedTag: "",
		},
		{
			name:        "tagged digest reference preserves tag",
			image:       "nginx:1.25@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expectedTag: "1.25",
		},
		{
			name:        "digest algorithm as repository name",
			image:       "sha256:v1",
			expectedTag: "v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := metadata(container.Summary{
				Image: tc.image,
			})
			assert.Equal(t, tc.image, got["ctn_image"])
			if tc.expectedTag != "" {
				assert.Equal(t, tc.expectedTag, got["ctn_image_tag"])
			} else {
				assert.NotContains(t, got, "ctn_image_tag")
			}
		})
	}
}

func TestFormatSizeOmitsVirtualSizeWhenMissing(t *testing.T) {
	assert.Equal(t, "1.02kB", formatSize(1024, 0))
}
