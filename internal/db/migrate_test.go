package db

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

const v1ManifestFixture = `{
	"Name": "docker.io/library/alpine",
	"Tag": "3.20",
	"MIMEType": "application/vnd.docker.distribution.manifest.v2+json",
	"Digest": "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	"Created": "2026-05-24T00:00:00Z",
	"DockerVersion": "25.0.0",
	"Labels": {
		"org.opencontainers.image.title": "alpine"
	},
	"Architecture": "amd64",
	"Os": "linux",
	"Layers": [
		"sha256:abcdef"
	]
}`

func TestMigrateV1ToV2(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "diun.db")
	createV1Database(t, dbPath)

	client, err := New(model.Db{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	require.NoError(t, client.Migrate())

	assert.Equal(t, Metadata{Version: 2}, client.metadata)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		entry := tx.Bucket([]byte(bucketManifest)).Get([]byte("docker.io/library/alpine:3.20"))
		require.NotNil(t, entry)

		var migrated map[string]any
		require.NoError(t, json.Unmarshal(entry, &migrated))
		assert.NotContains(t, migrated, "Architecture")
		assert.NotContains(t, migrated, "Os")
		assert.Equal(t, "docker.io/library/alpine", migrated["Name"])
		assert.Equal(t, "3.20", migrated["Tag"])
		return nil
	}))
}

func createV1Database(t *testing.T, dbPath string) {
	t.Helper()

	rawDB, err := bolt.Open(dbPath, 0600, &bolt.Options{
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		if rawDB != nil {
			_ = rawDB.Close()
		}
	}()

	require.NoError(t, rawDB.Update(func(tx *bolt.Tx) error {
		metadataBucket, err := tx.CreateBucketIfNotExists([]byte(bucketMetadata))
		if err != nil {
			return err
		}
		if err := metadataBucket.Put([]byte(metadataKey), []byte(`{"Version":1}`)); err != nil {
			return err
		}
		manifestBucket, err := tx.CreateBucketIfNotExists([]byte(bucketManifest))
		if err != nil {
			return err
		}
		return manifestBucket.Put([]byte("docker.io/library/alpine:3.20"), []byte(v1ManifestFixture))
	}))

	require.NoError(t, rawDB.Close())
	rawDB = nil
}
