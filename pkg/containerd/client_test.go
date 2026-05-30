package containerd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestNormalizeEndpoint(t *testing.T) {
	assert.Equal(t, "/run/containerd/containerd.sock", normalizeEndpoint("unix:///run/containerd/containerd.sock"))
	assert.Equal(t, `//./pipe/containerd-containerd`, normalizeEndpoint(`npipe:////./pipe/containerd-containerd`))
	assert.Equal(t, "/run/containerd/containerd.sock", normalizeEndpoint("/run/containerd/containerd.sock"))
}

func TestWithNamespace(t *testing.T) {
	ctx := withNamespace(context.Background(), "default")
	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	assert.Equal(t, []string{"default"}, md.Get(namespaceHeader))
}
