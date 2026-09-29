package docker

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

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
		name       string
		image      string
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
