package db

import (
	"sync"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestManifestStore(t *testing.T) {
	client := newTestClient(t)
	alpine := parseTestImage(t, "alpine:3.20")
	alpine2 := parseTestImage(t, "alpine2:3.20")
	alpineManifest := testManifest("docker.io/library/alpine", "3.20")
	alpine2Manifest := testManifest("docker.io/library/alpine2", "3.20")

	first, err := client.First(alpine)
	require.NoError(t, err)
	assert.True(t, first)

	require.NoError(t, client.PutManifest(alpine2, alpine2Manifest))
	first, err = client.First(alpine)
	require.NoError(t, err)
	assert.True(t, first)

	require.NoError(t, client.PutManifest(alpine, alpineManifest))
	first, err = client.First(alpine)
	require.NoError(t, err)
	assert.False(t, first)

	stored, err := client.GetManifest(alpine)
	require.NoError(t, err)
	assert.Equal(t, alpineManifest, stored)

	images, err := client.ListImage()
	require.NoError(t, err)
	assert.Equal(t, map[string][]ManifestEntry{
		"docker.io/library/alpine": {
			{
				Reference: alpine.String(),
				Manifest:  alpineManifest,
			},
		},
		"docker.io/library/alpine2": {
			{
				Reference: alpine2.String(),
				Manifest:  alpine2Manifest,
			},
		},
	}, images)

	manifests, err := client.ListManifest()
	require.NoError(t, err)
	assert.ElementsMatch(t, []registry.Manifest{alpineManifest, alpine2Manifest}, manifests)
}

func TestPutManifestMarshalErrorPreservesHistory(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:3.20")
	manifest := testManifest(image.Name(), image.Tag)
	require.NoError(t, client.PutManifest(image, manifest))

	invalid := manifest
	invalid.Created = new(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC))
	require.Error(t, client.PutManifest(image, invalid))

	stored, err := client.GetManifest(image)
	require.NoError(t, err)
	require.Equal(t, manifest, stored)
}

func TestDeleteManifestRemovesTaggedDigestEntry(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:3.20@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	manifest := testManifest("docker.io/library/alpine", "3.20")
	manifest.Digest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	require.NoError(t, client.PutManifest(image, manifest))
	_, err := client.DeleteManifest(image.String())
	require.NoError(t, err)

	manifests, err := client.ListManifest()
	require.NoError(t, err)
	assert.Empty(t, manifests)
}

func parseTestImage(t *testing.T, name string) registry.Image {
	t.Helper()
	image, err := registry.ParseImage(registry.ParseImageOptions{Name: name})
	require.NoError(t, err)
	return image
}

func testManifest(name, tag string) registry.Manifest {
	return registry.Manifest{
		Name:     name,
		Tag:      tag,
		MIMEType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:   digest.FromString(name + ":" + tag),
		Created:  new(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)),
		Platform: "linux/amd64",
	}
}

func TestManifestStoreDigestReference(t *testing.T) {
	client := newTestClient(t)
	dgst := digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	alpine := parseTestImage(t, "alpine@"+dgst.String())
	alpine2 := parseTestImage(t, "alpine2@"+dgst.String())
	alpineManifest := testManifest("docker.io/library/alpine", "")
	alpineManifest.Digest = dgst
	alpine2Manifest := testManifest("docker.io/library/alpine2", "")
	alpine2Manifest.Digest = dgst

	first, err := client.First(alpine)
	require.NoError(t, err)
	assert.True(t, first)

	// another repository sharing the same prefix must not be seen as a match
	require.NoError(t, client.PutManifest(alpine2, alpine2Manifest))
	first, err = client.First(alpine)
	require.NoError(t, err)
	assert.True(t, first)

	require.NoError(t, client.PutManifest(alpine, alpineManifest))
	first, err = client.First(alpine)
	require.NoError(t, err)
	assert.False(t, first)

	stored, err := client.GetManifest(alpine)
	require.NoError(t, err)
	assert.Equal(t, alpineManifest, stored)
}

func TestFirstMatchesTagAndDigestReferences(t *testing.T) {
	client := newTestClient(t)
	dgst := digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	tagged := parseTestImage(t, "alpine:3.20")
	digested := parseTestImage(t, "alpine@"+dgst.String())
	manifest := testManifest("docker.io/library/alpine", "3.20")
	manifest.Digest = dgst

	// an image already analyzed through its tag is not a first check when
	// referenced by digest, and the other way around
	require.NoError(t, client.PutManifest(tagged, manifest))
	first, err := client.First(digested)
	require.NoError(t, err)
	assert.False(t, first)

	client2 := newTestClient(t)
	manifest.Tag = ""
	require.NoError(t, client2.PutManifest(digested, manifest))
	first, err = client2.First(tagged)
	require.NoError(t, err)
	assert.False(t, first)
}

func TestDeleteManifestRemovesDigestOnlyEntry(t *testing.T) {
	client := newTestClient(t)
	dgst := digest.Digest("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	image := parseTestImage(t, "alpine@"+dgst.String())
	manifest := testManifest("docker.io/library/alpine", "")
	manifest.Digest = dgst

	require.NoError(t, client.PutManifest(image, manifest))
	_, err := client.DeleteManifest(image.String())
	require.NoError(t, err)

	manifests, err := client.ListManifest()
	require.NoError(t, err)
	assert.Empty(t, manifests)
}

func TestDeleteManifestPreservesDigestOnlyEntry(t *testing.T) {
	client := newTestClient(t)
	dgst := digest.FromString("shared-content")
	tagged := parseTestImage(t, "alpine:3.20")
	pinned := parseTestImage(t, "alpine@"+dgst.String())
	taggedManifest := testManifest(tagged.Name(), tagged.Tag)
	taggedManifest.Digest = dgst
	pinnedManifest := testManifest(pinned.Name(), "")
	pinnedManifest.Digest = dgst
	require.NoError(t, client.PutManifest(tagged, taggedManifest))
	require.NoError(t, client.PutManifest(pinned, pinnedManifest))
	_, err := client.DeleteManifest(tagged.String())
	require.NoError(t, err)
	stored, err := client.GetManifest(pinned)
	require.NoError(t, err)
	assert.Equal(t, pinnedManifest, stored, "removing a tag should preserve the separately stored digest-only reference")
}

func TestPlatformManifestsKeepIndependentBaselines(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	amd := model.ImagePlatform{OS: "linux", Arch: "amd64"}
	arm := model.ImagePlatform{OS: "linux", Arch: "arm64"}
	old := testManifest(image.Name(), image.Tag)
	require.NoError(t, client.PutManifest(image, old))
	baseline, err := client.GetPlatformManifest(image, arm)
	require.NoError(t, err)
	require.Equal(t, old, baseline, "new platform inherits reference history")

	amdManifest, armManifest := old, old
	amdManifest.InstanceDigest = digest.FromString("amd64")
	armManifest.InstanceDigest = digest.FromString("arm64")
	armManifest.Platform = "linux/arm64"
	var wg sync.WaitGroup
	wg.Go(func() { assert.NoError(t, client.PutPlatformManifest(image, amd, amdManifest)) })
	wg.Go(func() { assert.NoError(t, client.PutPlatformManifest(image, arm, armManifest)) })
	wg.Wait()
	stored, err := client.GetPlatformManifest(image, amd)
	require.NoError(t, err)
	require.Equal(t, amdManifest, stored)
	stored, err = client.GetPlatformManifest(image, arm)
	require.NoError(t, err)
	require.Equal(t, armManifest, stored)

	updated := amdManifest
	updated.Digest = digest.FromString("new index")
	require.NoError(t, client.PutPlatformManifest(image, amd, updated))
	stored, err = client.GetPlatformManifest(image, arm)
	require.NoError(t, err)
	require.Equal(t, armManifest, stored, "another platform must not advance this baseline")
	first, err := client.First(image)
	require.NoError(t, err)
	require.False(t, first)
	listed, err := client.ListManifest()
	require.NoError(t, err)
	require.Equal(t, []registry.Manifest{updated}, listed)
	images, err := client.ListImage()
	require.NoError(t, err)
	require.Len(t, images[image.Name()], 1)
}

func TestPlatformManifestKeysPreserveSelectionSettings(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	platforms := []model.ImagePlatform{
		{},
		{
			Arch: "amd64",
		},
		{
			Arch: "arm64",
		},
		{
			OS:   "linux",
			Arch: "arm",
		},
		{
			OS:      "linux",
			Arch:    "arm",
			Variant: "v6",
		},
		{
			OS:      "linux",
			Arch:    "arm",
			Variant: "v7",
		},
	}
	for i, platform := range platforms {
		manifest := testManifest(image.Name(), image.Tag)
		manifest.Digest = digest.FromBytes([]byte{byte(i)})
		require.NoError(t, client.PutPlatformManifest(image, platform, manifest))
	}
	for i, platform := range platforms {
		stored, err := client.GetPlatformManifest(image, platform)
		require.NoError(t, err)
		require.Equal(t, digest.FromBytes([]byte{byte(i)}), stored.Digest)
	}
}

func TestDeleteManifestClearsAllPlatformRecords(t *testing.T) {
	for _, ref := range []string{"alpine:latest", "alpine@" + digest.FromString("pinned").String(), "alpine:latest@" + digest.FromString("pinned").String()} {
		t.Run(ref, func(t *testing.T) {
			client := newTestClient(t)
			image := parseTestImage(t, ref)
			other := parseTestImage(t, "alpine:latest-other")
			manifest := testManifest(image.Name(), image.Tag)
			manifest.Digest = digest.FromString("tag moved")
			for _, arch := range []string{"amd64", "arm64"} {
				platform := model.ImagePlatform{OS: "linux", Arch: arch}
				require.NoError(t, client.PutPlatformManifest(image, platform, manifest))
				require.NoError(t, client.PutPlatformManifest(other, platform, testManifest(other.Name(), other.Tag)))
			}
			_, err := client.DeleteManifest(image.String())
			require.NoError(t, err)
			size, err := client.DeleteManifest(image.String())
			require.NoError(t, err, "removal is idempotent")
			require.Zero(t, size)
			require.NoError(t, client.View(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(bucketManifestPlatformState))
				require.Nil(t, bucket.Bucket([]byte(image.String())))
				require.NotNil(t, bucket.Bucket([]byte(other.String())))
				return nil
			}))
			stored, err := client.GetPlatformManifest(image, model.ImagePlatform{OS: "linux", Arch: "amd64"})
			require.NoError(t, err)
			require.Empty(t, stored)
		})
	}
}

func TestPlatformManifestWriteIsAtomic(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	platform := model.ImagePlatform{OS: "linux", Arch: "amd64"}
	manifest := testManifest(image.Name(), image.Tag)
	// A bucket at the reference key makes the history write fail after the
	// platform write. The transaction must roll back the platform write too.
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		_, err := tx.Bucket([]byte(bucketManifest)).CreateBucket([]byte(image.String()))
		return err
	}))
	require.Error(t, client.PutPlatformManifest(image, platform, manifest))
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		require.Nil(t, tx.Bucket([]byte(bucketManifestPlatformState)).Bucket([]byte(image.String())))
		return nil
	}))
}

func TestDeleteManifestCountsStoredValues(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	manifest := testManifest(image.Name(), image.Tag)
	// Preserve an unknown field and whitespace to verify that size is measured
	// from stored bytes, not by re-encoding the caller's manifest.
	const history = ` {"Name":"docker.io/library/alpine","Tag":"latest","extra":true} `
	const state = `{"Platform":"linux/amd64"}`
	const pinned = `{"Platform":"linux/arm64"}`
	taggedDigest := parseTestImage(t, image.String()+"@"+manifest.Digest.String())
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketManifest))
		if err := bucket.Put([]byte(image.String()), []byte(history)); err != nil {
			return err
		}
		if err := bucket.Put([]byte(taggedDigest.String()), []byte(pinned)); err != nil {
			return err
		}
		platforms, err := tx.Bucket([]byte(bucketManifestPlatformState)).CreateBucket([]byte(image.String()))
		if err != nil {
			return err
		}
		return platforms.Put(platformSettingsKey(model.ImagePlatform{OS: "linux", Arch: "amd64"}), []byte(state))
	}))
	size, err := client.DeleteManifest(image.String())
	require.NoError(t, err)
	require.Equal(t, int64(len(history)+len(state)), size)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		require.JSONEq(t, pinned, string(tx.Bucket([]byte(bucketManifest)).Get([]byte(taggedDigest.String()))))
		return nil
	}))
	size, err = client.DeleteManifest(image.String())
	require.NoError(t, err)
	require.Zero(t, size)
}

func TestDeleteManifestFailureReturnsNoSizeAndPreservesPlatformState(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:latest")
	manifest := testManifest(image.Name(), image.Tag)
	platform := model.ImagePlatform{OS: "linux", Arch: "amd64"}
	require.NoError(t, client.PutPlatformManifest(image, platform, manifest))
	// A bucket at the history key makes deletion fail. Platform state must
	// remain intact, and no removed bytes may be reported.
	require.NoError(t, client.Update(func(tx *bolt.Tx) error {
		history := tx.Bucket([]byte(bucketManifest))
		if err := history.Delete([]byte(image.String())); err != nil {
			return err
		}
		_, err := history.CreateBucket([]byte(image.String()))
		return err
	}))
	size, err := client.DeleteManifest(image.String())
	require.Error(t, err)
	require.Zero(t, size)
	require.NoError(t, client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketManifestPlatformState)).Bucket([]byte(image.String()))
		require.NotNil(t, bucket)
		require.NotNil(t, bucket.Get(platformSettingsKey(platform)))
		return nil
	}))
}

func TestManifestEntryPreservesStoredReference(t *testing.T) {
	client := newTestClient(t)
	old := digest.FromString("deployed")
	current := digest.FromString("registry now")
	refs := []string{
		"alpine:latest",
		"alpine:latest@" + old.String(),
		"alpine@" + old.String(),
	}
	// All references resolve to the same current tag. Their original database
	// keys must remain distinct even though the observed metadata is identical.
	manifest := testManifest("docker.io/library/alpine", "latest")
	manifest.Digest = current
	var want []ManifestEntry
	for _, ref := range refs {
		image := parseTestImage(t, ref)
		require.NoError(t, client.PutManifest(image, manifest))
		want = append(want, ManifestEntry{Reference: image.String(), Manifest: manifest})
	}
	images, err := client.ListImage()
	require.NoError(t, err)
	require.ElementsMatch(t, want, images[manifest.Name])

	for i, entry := range want {
		_, err := client.DeleteManifest(entry.Reference)
		require.NoError(t, err)
		_, err = client.DeleteManifest(entry.Reference)
		require.NoError(t, err, "deleting a missing reference is harmless")
		images, err = client.ListImage()
		require.NoError(t, err)
		require.ElementsMatch(t, want[i+1:], images[manifest.Name], "deletion must preserve other references with identical observations")
	}
	first, err := client.First(parseTestImage(t, "alpine:latest"))
	require.NoError(t, err)
	require.True(t, first)
}
