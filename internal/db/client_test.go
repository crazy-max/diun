package db

import (
	"path/filepath"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestNewInitializesBucketsAndMetadata(t *testing.T) {
	client := newTestClient(t)
	assert.Equal(t, Metadata{Version: 1}, client.metadata)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		assert.NotNil(t, tx.Bucket([]byte(bucketMetadata)))
		assert.NotNil(t, tx.Bucket([]byte(bucketManifest)))
		return nil
	}))
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(model.Db{Path: filepath.Join(t.TempDir(), "diun.db")})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
	return client
}
