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
		"ctn_command":   "nginx -g daemon off;",
		"ctn_createdat": time.Unix(created, 0).String(),
		"ctn_state":     "running",
		"ctn_status":    "Up 5 minutes",
		"ctn_size":      "1.02kB (virtual 2.05kB)",
	}, got)
}

func TestFormatSizeOmitsVirtualSizeWhenMissing(t *testing.T) {
	assert.Equal(t, "1.02kB", formatSize(1024, 0))
}
