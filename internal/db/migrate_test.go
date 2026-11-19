package db

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/opencontainers/go-digest"
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

func TestMigrateV1ToLatest(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "diun.db")
	createV1Database(t, dbPath)

	client, err := New(model.Db{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	require.NoError(t, client.Migrate())

	assert.Equal(t, Metadata{Version: 3}, client.metadata)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		entry := tx.Bucket([]byte(bucketManifest)).Get([]byte("docker.io/library/alpine:3.20"))
		require.NotNil(t, entry)

		var migrated map[string]any
		require.NoError(t, json.Unmarshal(entry, &migrated))
		assert.NotContains(t, migrated, "Architecture")
		assert.NotContains(t, migrated, "Os")
		assert.Equal(t, "docker.io/library/alpine", migrated["Name"])
		assert.Equal(t, "3.20", migrated["Tag"])
		var original map[string]any
		require.NoError(t, json.Unmarshal([]byte(v1ManifestFixture), &original))
		delete(original, "Architecture")
		delete(original, "Os")
		assert.Equal(t, original, migrated, "migration must preserve all other fields")
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

func TestMigrateV2PreservesHistoryAndResetsPlatformCache(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	removed := parseTestImage(t, "alpine:removed")
	platform := model.ImagePlatform{OS: "linux", Arch: "arm64"}
	old := testManifest(image.Name(), image.Tag)
	old.InstanceDigest = digest.FromString("arm64")
	require.NoError(t, client.PutPlatformManifest(image, platform, old))
	require.NoError(t, client.PutPlatformManifest(removed, platform, old))
	// Simulate a v2 writer: it updates reference history and does not know
	// about platform records left by a newer binary.
	latest := old
	latest.Digest = digest.FromString("new history")
	require.NoError(t, client.PutManifest(image, latest))
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketManifest)).Delete([]byte(removed.String()))
	}))
	require.NoError(t, client.WriteMetadata(Metadata{Version: 2}))
	require.NoError(t, client.Migrate())
	require.Equal(t, 3, client.metadata.Version)
	stored, err := client.GetPlatformManifest(image, platform)
	require.NoError(t, err)
	require.Equal(t, latest, stored)
	stored, err = client.GetPlatformManifest(removed, platform)
	require.NoError(t, err)
	require.Empty(t, stored)
	first, err := client.First(image)
	require.NoError(t, err)
	require.False(t, first)
	// Re-running migration at v3 must not discard platform-specific baselines.
	require.NoError(t, client.PutPlatformManifest(image, platform, old))
	require.NoError(t, client.PutManifest(image, latest))
	require.NoError(t, client.Migrate())
	stored, err = client.GetPlatformManifest(image, platform)
	require.NoError(t, err)
	require.Equal(t, old, stored)
}

func TestMigrateEmptyDatabase(t *testing.T) {
	client := newTestClient(t)
	require.NoError(t, client.Migrate())
	require.NoError(t, client.ReadMetadata())
	require.Equal(t, dbVersion, client.metadata.Version)
	manifests, err := client.ListManifest()
	require.NoError(t, err)
	require.Empty(t, manifests)
	first, err := client.First(parseTestImage(t, "alpine:latest"))
	require.NoError(t, err)
	require.True(t, first)
}

func TestMigrateV2WithoutPlatformState(t *testing.T) {
	client := newTestClient(t)
	refs := []string{
		"alpine:latest",
		"alpine@" + digest.FromString("pinned").String(),
		"alpine:latest@" + digest.FromString("pinned").String(),
	}
	entries := make(map[string][]byte)
	for _, ref := range refs {
		image := parseTestImage(t, ref)
		manifest := testManifest(image.Name(), image.Tag)
		if image.Digest != "" {
			manifest.Digest = image.Digest
		}
		entry, err := json.Marshal(manifest)
		require.NoError(t, err)
		entries[image.String()] = entry
		require.NoError(t, client.PutManifest(image, manifest))
	}
	require.NoError(t, client.WriteMetadata(Metadata{Version: 2}))
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketManifestPlatformState))
	}))

	require.NoError(t, client.Migrate())
	require.NoError(t, client.ReadMetadata())
	require.Equal(t, dbVersion, client.metadata.Version)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		for key, entry := range entries {
			require.Equal(t, entry, tx.Bucket([]byte(bucketManifest)).Get([]byte(key)))
		}
		platforms := tx.Bucket([]byte(bucketManifestPlatformState))
		require.NotNil(t, platforms)
		key, _ := platforms.Cursor().First()
		require.Nil(t, key, "platform state is populated lazily")
		return nil
	}))
}

func TestMigrateV1FailureRollsBackAndCanRetry(t *testing.T) {
	client := newTestClient(t)
	const validKey = "docker.io/library/alpine:3.20"
	const invalidKey = "docker.io/library/busybox:latest"
	const invalidJSON = `{"Name":`
	require.NoError(t, client.WriteMetadata(Metadata{Version: 1}))
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketManifest))
		// The valid entry sorts first, so migration rewrites it before failing.
		if err := bucket.Put([]byte(validKey), []byte(v1ManifestFixture)); err != nil {
			return err
		}
		return bucket.Put([]byte(invalidKey), []byte(invalidJSON))
	}))

	err := client.Migrate()
	require.ErrorContains(t, err, "database migration v2 failed")
	var syntaxError *json.SyntaxError
	require.ErrorAs(t, err, &syntaxError)
	require.Equal(t, 1, client.metadata.Version)
	require.NoError(t, client.ReadMetadata())
	require.Equal(t, 1, client.metadata.Version, "failed migration must not advance the stored version")
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketManifest))
		require.True(t, bytes.Equal([]byte(v1ManifestFixture), bucket.Get([]byte(validKey))), "valid record must remain byte-for-byte intact")
		require.True(t, bytes.Equal([]byte(invalidJSON), bucket.Get([]byte(invalidKey))), "invalid record must remain byte-for-byte intact")
		return nil
	}))

	// Remove the corrupt record and retry on the same connection. This also
	// verifies that the failed migration released its write transaction.
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketManifest)).Delete([]byte(invalidKey))
	}))
	require.NoError(t, client.Migrate())
	require.NoError(t, client.ReadMetadata())
	require.Equal(t, dbVersion, client.metadata.Version)
	stored, err := client.GetManifest(parseTestImage(t, "alpine:3.20"))
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/alpine", stored.Name)
	require.Equal(t, "3.20", stored.Tag)
}

func TestPlatformCacheSurvivesReopen(t *testing.T) {
	cfg := model.Db{Path: filepath.Join(t.TempDir(), "diun.db")}
	client, err := New(cfg)
	require.NoError(t, err)
	defer func() { require.NoError(t, client.Close()) }()
	require.NoError(t, client.Migrate())
	image := parseTestImage(t, "alpine:latest")
	amd := model.ImagePlatform{OS: "linux", Arch: "amd64"}
	arm := model.ImagePlatform{OS: "linux", Arch: "arm64"}
	first := testManifest(image.Name(), image.Tag)
	second := first
	second.Platform = "linux/arm64"
	require.NoError(t, client.PutPlatformManifest(image, amd, first))
	require.NoError(t, client.PutPlatformManifest(image, arm, second))
	require.NoError(t, client.Close())
	client, err = New(cfg)
	require.NoError(t, err)
	require.NoError(t, client.Migrate())
	stored, err := client.GetPlatformManifest(image, amd)
	require.NoError(t, err)
	require.Equal(t, first, stored)
	stored, err = client.GetPlatformManifest(image, arm)
	require.NoError(t, err)
	require.Equal(t, second, stored)
}
