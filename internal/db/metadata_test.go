package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	client := newTestClient(t)
	require.NoError(t, client.WriteMetadata(Metadata{Version: 2}))
	client.metadata = Metadata{}
	require.NoError(t, client.ReadMetadata())
	assert.Equal(t, Metadata{Version: 2}, client.metadata)
}
