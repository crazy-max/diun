package db

import (
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, map[string][]registry.Manifest{
		"docker.io/library/alpine":  {alpineManifest},
		"docker.io/library/alpine2": {alpine2Manifest},
	}, images)

	manifests, err := client.ListManifest()
	require.NoError(t, err)
	assert.ElementsMatch(t, []registry.Manifest{alpineManifest, alpine2Manifest}, manifests)
}

func TestDeleteManifestRemovesTaggedDigestEntry(t *testing.T) {
	client := newTestClient(t)
	image := parseTestImage(t, "alpine:3.20@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	manifest := testManifest("docker.io/library/alpine", "3.20")
	manifest.Digest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	require.NoError(t, client.PutManifest(image, manifest))
	require.NoError(t, client.DeleteManifest(manifest))

	manifests, err := client.ListManifest()
	require.NoError(t, err)
	assert.Empty(t, manifests)
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
	require.NoError(t, client.DeleteManifest(manifest))

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
