package docker

import (
	"testing"

	"github.com/moby/moby/api/types/image"
	"github.com/stretchr/testify/assert"
)

func TestIsDigest(t *testing.T) {
	c := &Client{}

	tests := []struct {
		name    string
		imageID string
		want    bool
	}{
		{
			name:    "bare digest",
			imageID: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			want:    true,
		},
		{
			name:    "at-prefixed digest",
			imageID: "@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			want:    true,
		},
		{
			name:    "at-prefixed hex",
			imageID: "@0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			want:    true,
		},
		{
			name:    "named digest reference",
			imageID: "alpine@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
		{
			name:    "short digest",
			imageID: "sha256:0123456789abcdef",
		},
		{
			name:    "uppercase digest",
			imageID: "sha256:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, c.IsDigest(tt.imageID))
		})
	}
}

func TestIsLocalImage(t *testing.T) {
	c := &Client{}

	assert.True(t, c.IsLocalImage(image.InspectResponse{}))
	assert.False(t, c.IsLocalImage(image.InspectResponse{
		RepoDigests: []string{"alpine@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}))
}

func TestIsDanglingImage(t *testing.T) {
	c := &Client{}

	assert.True(t, c.IsDanglingImage(image.InspectResponse{
		RepoTags:    []string{"<none>:<none>"},
		RepoDigests: []string{"<none>@<none>"},
	}))
	assert.False(t, c.IsDanglingImage(image.InspectResponse{
		RepoTags:    []string{"alpine:latest"},
		RepoDigests: []string{"alpine@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}))
}
