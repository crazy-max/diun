package registry

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"testing"

	digest "github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	podmanmanifest "go.podman.io/image/v5/manifest"
)

func TestManifestReturnsCachedManifestWhenDigestMatches(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, imgspecv1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, "25.0.0", nil)
	registry.addImage("1.0.0", remoteImage)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.0.0"),
	})
	require.NoError(t, err)

	cachedManifest := Manifest{
		Name:     "cached.example/acme/diun",
		Tag:      "1.0.0",
		MIMEType: "application/vnd.test.cached",
		Digest:   remoteImage.manifest.digest,
		Platform: "linux/amd64",
	}

	client := newTestRegistryClient(t, Options{CompareDigest: true})
	manifest, updated, err := client.Manifest(image, cachedManifest)
	require.NoError(t, err)

	assert.False(t, updated)
	assert.Equal(t, cachedManifest, manifest)
	assert.Equal(t, 1, registry.requestCount(http.MethodHead, "/v2/acme/diun/manifests/1.0.0"))
	assert.Zero(t, registry.requestCount(http.MethodGet, "/v2/acme/diun/manifests/1.0.0"))
}

func TestManifest(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, imgspecv1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, "25.0.0", map[string]string{
		"org.opencontainers.image.title": "diun",
	})
	registry.addImage("1.0.0", remoteImage)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.0.0"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{CompareDigest: true})
	manifest, updated, err := client.Manifest(image, Manifest{
		Digest: digest.FromString("stale manifest"),
	})
	require.NoError(t, err)

	assert.True(t, updated)
	assert.Equal(t, registry.host()+"/acme/diun", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Tag)
	assert.Equal(t, podmanmanifest.DockerV2Schema2MediaType, manifest.MIMEType)
	assert.Equal(t, remoteImage.manifest.digest, manifest.Digest)
	assert.Equal(t, "25.0.0", manifest.DockerVersion)
	assert.Equal(t, map[string]string{
		"org.opencontainers.image.title": "diun",
	}, manifest.Labels)
	assert.Equal(t, []string{remoteImage.layer.String()}, manifest.Layers)
	assert.Equal(t, "linux/amd64", manifest.Platform)
	assert.NotEmpty(t, manifest.Raw)
}

func TestManifestListPlatformDigestComparison(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	amd64Platform := imgspecv1.Platform{Architecture: "amd64", OS: "linux"}
	armPlatform := imgspecv1.Platform{Architecture: "arm", OS: "linux", Variant: "v7"}

	remoteAMD64 := newTestRegistryImage(t, amd64Platform, "25.0.0", nil)
	remoteARM := newTestRegistryImage(t, armPlatform, "25.0.0", nil)
	oldAMD64 := newTestRegistryImage(t, amd64Platform, "24.0.0", nil)
	oldARM := newTestRegistryImage(t, armPlatform, "24.0.0", nil)

	remoteList := newTestManifestList(t,
		testManifestListInstance{manifest: remoteAMD64.manifest, platform: amd64Platform},
		testManifestListInstance{manifest: remoteARM.manifest, platform: armPlatform},
	)
	dbListWithSamePlatform := newTestManifestList(t,
		testManifestListInstance{manifest: remoteAMD64.manifest, platform: amd64Platform},
		testManifestListInstance{manifest: oldARM.manifest, platform: armPlatform},
	)
	dbListWithOldPlatform := newTestManifestList(t,
		testManifestListInstance{manifest: oldAMD64.manifest, platform: amd64Platform},
		testManifestListInstance{manifest: remoteARM.manifest, platform: armPlatform},
	)

	registry.addImage("linux-amd64", remoteAMD64)
	registry.addImage("linux-arm-v7", remoteARM)
	registry.addManifest("multi", remoteList)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("multi"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})

	for _, tt := range []struct {
		name    string
		raw     []byte
		digest  digest.Digest
		updated bool
	}{
		{
			name:   "same selected platform digest",
			raw:    dbListWithSamePlatform.body,
			digest: dbListWithSamePlatform.digest,
		},
		{
			name:    "different selected platform digest",
			raw:     dbListWithOldPlatform.body,
			digest:  dbListWithOldPlatform.digest,
			updated: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manifest, updated, err := client.Manifest(image, Manifest{
				MIMEType: podmanmanifest.DockerV2ListMediaType,
				Digest:   tt.digest,
				Raw:      tt.raw,
			})
			require.NoError(t, err)

			assert.Equal(t, tt.updated, updated)
			assert.Equal(t, registry.host()+"/acme/diun", manifest.Name)
			assert.Equal(t, "multi", manifest.Tag)
			assert.Equal(t, podmanmanifest.DockerV2ListMediaType, manifest.MIMEType)
			assert.Equal(t, remoteList.digest, manifest.Digest)
			assert.Equal(t, "linux/amd64", manifest.Platform)
		})
	}
}

func TestManifestVariant(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	amd64Platform := imgspecv1.Platform{Architecture: "amd64", OS: "linux"}
	armPlatform := imgspecv1.Platform{Architecture: "arm", OS: "linux", Variant: "v7"}

	remoteAMD64 := newTestRegistryImage(t, amd64Platform, "25.0.0", nil)
	remoteARM := newTestRegistryImage(t, armPlatform, "25.0.0", nil)
	remoteList := newTestManifestList(t,
		testManifestListInstance{manifest: remoteAMD64.manifest, platform: amd64Platform},
		testManifestListInstance{manifest: remoteARM.manifest, platform: armPlatform},
	)

	registry.addImage("linux-amd64", remoteAMD64)
	registry.addImage("linux-arm-v7", remoteARM)
	registry.addManifest("multi", remoteList)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("multi"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{
		ImageOs:      "linux",
		ImageArch:    "arm",
		ImageVariant: "v7",
	})
	manifest, updated, err := client.Manifest(image, Manifest{})
	require.NoError(t, err)

	assert.True(t, updated)
	assert.Equal(t, registry.host()+"/acme/diun", manifest.Name)
	assert.Equal(t, "multi", manifest.Tag)
	assert.Equal(t, "linux/arm/v7", manifest.Platform)
}

func TestManifestNonImageArtifact(t *testing.T) {
	const sigstoreBundleType = "application/vnd.dev.sigstore.bundle.v0.3+json"

	registry := newTestRegistry(t, "acme/diun")
	artifact := newTestOCIArtifactManifest(t, sigstoreBundleType)
	registry.addManifest("sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526", artifact)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{CompareDigest: true})
	_, _, err = client.Manifest(image, Manifest{})
	require.Error(t, err)

	_, ok := stderrors.AsType[podmanmanifest.NonImageArtifactError](err)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), "unsupported image-specific operation on artifact with type "+strconv.Quote(sigstoreBundleType))
}

func TestManifestTaggedDigestUnknownTag(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, imgspecv1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, "25.0.0", nil)
	registry.addImage("1.0.0", remoteImage)

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("missing") + "@" + remoteImage.manifest.digest.String(),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{CompareDigest: true})
	_, _, err = client.Manifest(image, Manifest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading digest missing")
}
