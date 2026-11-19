package registry

import (
	"encoding/json"
	stderrors "errors"
	"math"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	digest "github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/regclient/regclient/types/docker/schema1"
	"github.com/regclient/regclient/types/errs"
	regmanifest "github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	regplatform "github.com/regclient/regclient/types/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestHeadRetry(t *testing.T) {
	for _, tt := range []struct {
		name      string
		statuses  []int
		wantHeads int
		wantError bool
		minDelay  time.Duration
	}{
		{
			name:      "internal server error",
			statuses:  []int{http.StatusInternalServerError},
			wantHeads: 2,
			minDelay:  100 * time.Millisecond,
		},
		{
			name:      "bad gateway",
			statuses:  []int{http.StatusBadGateway},
			wantHeads: 2,
			minDelay:  100 * time.Millisecond,
		},
		{
			// TODO: Migration blocker: regclient does not retry 503 responses.
			name:      "service unavailable is not retried upstream",
			statuses:  []int{http.StatusServiceUnavailable},
			wantHeads: 1,
			wantError: true,
		},
		{
			name:      "gateway timeout",
			statuses:  []int{http.StatusGatewayTimeout},
			wantHeads: 2,
			minDelay:  100 * time.Millisecond,
		},
		{
			name:      "success on last attempt",
			statuses:  []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusInternalServerError},
			wantHeads: 5,
			minDelay:  time.Second / 2,
		},
		{
			name:      "retry limit",
			statuses:  []int{http.StatusInternalServerError, http.StatusInternalServerError, http.StatusInternalServerError, http.StatusInternalServerError, http.StatusInternalServerError},
			wantHeads: 5,
			wantError: true,
			minDelay:  time.Second,
		},
		{
			name:      "unauthorized",
			statuses:  []int{http.StatusUnauthorized},
			wantHeads: 1,
			wantError: true,
		},
		{
			name:      "forbidden",
			statuses:  []int{http.StatusForbidden},
			wantHeads: 1,
			wantError: true,
		},
		{
			name:      "not found",
			statuses:  []int{http.StatusNotFound},
			wantHeads: 1,
			wantError: true,
		},
		{
			name:      "not implemented",
			statuses:  []int{http.StatusNotImplemented},
			wantHeads: 1,
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			registry := newTestRegistry(t, "acme/diun")
			registry.manifestStatuses = map[string][]int{http.MethodHead: tt.statuses}
			remoteImage := newTestRegistryImage(t, ocispecs.Platform{
				Architecture: "amd64",
				OS:           "linux",
			}, "25.0.0", nil)
			registry.addImage("1.0.0", remoteImage)
			image, err := ParseImage(ParseImageOptions{
				Name: registry.imageName("1.0.0"),
			})
			require.NoError(t, err)
			cachedManifest := Manifest{Digest: remoteImage.manifest.digest}
			client := newTestRegistryClient(t, Options{CompareDigest: true})

			start := time.Now()
			manifest, updated, err := client.Manifest(image, cachedManifest)
			assert.GreaterOrEqual(t, time.Since(start), tt.minDelay)
			if tt.wantError {
				require.ErrorContains(t, err, "cannot get image digest from HEAD request")
				assert.Empty(t, manifest)
				if tt.wantHeads > 1 {
					require.ErrorIs(t, err, errs.ErrHTTPStatus)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, cachedManifest, manifest)
			}
			assert.False(t, updated)
			assert.Equal(t, tt.wantHeads, registry.requestCount(http.MethodHead, "/v2/acme/diun/manifests/1.0.0"))
			assert.Zero(t, registry.requestCount(http.MethodGet, "/v2/acme/diun/manifests/1.0.0"))
		})
	}
}

func TestManifestRetriesGet(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	registry.manifestStatuses = map[string][]int{
		http.MethodGet: {http.StatusBadGateway},
	}
	remoteImage := newTestRegistryImage(t, ocispecs.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, "25.0.0", nil)
	registry.addImage("1.0.0", remoteImage)
	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.0.0"),
	})
	require.NoError(t, err)
	client := newTestRegistryClient(t, Options{})

	manifest, updated, err := client.Manifest(image, Manifest{})
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, remoteImage.manifest.digest, manifest.Digest)
	assert.Equal(t, "linux/amd64", manifest.Platform)
	assert.Equal(t, 1, registry.requestCount(http.MethodHead, "/v2/acme/diun/manifests/1.0.0"))
	assert.Equal(t, 2, registry.requestCount(http.MethodGet, "/v2/acme/diun/manifests/"+remoteImage.manifest.digest.String()))
}

func TestManifestRejectsOverflowingConfigSize(t *testing.T) {
	for _, tt := range []struct {
		name         string
		manifestType string
		configType   string
	}{
		{
			name:         "docker",
			manifestType: mediatype.Docker2Manifest,
			configType:   mediatype.Docker2ImageConfig,
		},
		{
			name:         "oci",
			manifestType: mediatype.OCI1Manifest,
			configType:   mediatype.OCI1ImageConfig,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry(t, "acme/diun")
			remote := newTestRegistryImage(t, ocispecs.Platform{OS: "linux", Architecture: "amd64"}, "", nil)
			var m ocispecs.Manifest
			require.NoError(t, json.Unmarshal(remote.manifest.body, &m))
			m.MediaType = tt.manifestType
			m.Config.MediaType = tt.configType
			m.Config.Size = math.MaxInt64
			body, err := json.Marshal(m)
			require.NoError(t, err)
			remote.manifest = testRegistryBlob{mediaType: tt.manifestType, body: body, digest: digest.FromBytes(body)}
			r.addImage("latest", remote)
			// Chunked responses bypass Content-Length validation in regclient.
			r.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if strings.Contains(req.URL.Path, "/blobs/") {
					w.Header().Set("Content-Type", "application/octet-stream")
					w.(http.Flusher).Flush()
				}
				r.handle(w, req)
			})
			image, err := ParseImage(ParseImageOptions{Name: r.imageName("latest")})
			require.NoError(t, err)
			client := newTestRegistryClient(t, Options{})
			require.NotPanics(t, func() {
				m, updated, err := client.Manifest(image, Manifest{})
				require.ErrorContains(t, err, "invalid image config size")
				assert.Empty(t, m)
				assert.False(t, updated)
			})
			assert.Zero(t, r.requestCount(http.MethodGet, "/v2/acme/diun/blobs/"+remote.config.digest.String()))
		})
	}
}

func TestManifestReturnsCachedManifestWhenDigestMatches(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, ocispecs.Platform{
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

func TestManifestAliasesReuseImmutableManifests(t *testing.T) {
	r := newTestRegistry(t, "acme/diun")
	p := ocispecs.Platform{OS: "linux", Architecture: "amd64"}
	remote := newTestRegistryImage(t, p, "25.0.0", nil)
	index := newTestManifestList(t, testManifestListInstance{manifest: remote.manifest, platform: p})
	r.addImage("platform", remote)
	r.addManifest("latest", index)
	r.addManifest("1.0.0", index)
	c := newTestRegistryClient(t, Options{CompareDigest: true})
	for _, tag := range []string{"latest", "1.0.0"} {
		img, err := ParseImage(ParseImageOptions{Name: r.imageName(tag)})
		require.NoError(t, err)
		m, _, err := c.Manifest(img, Manifest{})
		require.NoError(t, err)
		assert.Equal(t, tag, m.Tag)
		_, changed, err := c.Manifest(img, m)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, 2, r.requestCount(http.MethodHead, "/v2/acme/diun/manifests/"+tag))
		assert.Zero(t, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/"+tag))
	}
	assert.Equal(t, 1, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/"+index.digest.String()))
	assert.Equal(t, 1, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/"+remote.manifest.digest.String()))
}

func TestManifestUsesRegclientPlatformSelection(t *testing.T) {
	for _, tt := range []struct {
		name      string
		opts      Options
		target    ocispecs.Platform
		wantError bool
	}{
		{
			name:   "Windows can select Linux",
			opts:   Options{Platform: regplatform.Platform{OS: "windows", Architecture: "amd64"}},
			target: ocispecs.Platform{OS: "linux", Architecture: "amd64"},
		},
		{
			name:      "Linux cannot select Windows",
			opts:      Options{Platform: regplatform.Platform{OS: "linux", Architecture: "amd64"}},
			target:    ocispecs.Platform{OS: "windows", Architecture: "amd64"},
			wantError: true,
		},
		{
			name:      "unspecified ARM variant defaults to v7",
			opts:      Options{Platform: regplatform.Platform{OS: "linux", Architecture: "arm"}},
			target:    ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v8"},
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry(t, "acme/diun")
			remote := newTestRegistryImage(t, tt.target, "", nil)
			r.addImage("target", remote)
			r.addManifest("multi", newTestManifestList(t, testManifestListInstance{manifest: remote.manifest, platform: tt.target}))
			image, err := ParseImage(ParseImageOptions{Name: r.imageName("multi")})
			require.NoError(t, err)
			client := newTestRegistryClient(t, tt.opts)
			_, _, err = client.Manifest(image, Manifest{})
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInspectSchema1(t *testing.T) {
	topLayer, baseLayer := digest.FromString("top"), digest.FromString("base")
	config := `{"created":"2026-01-01T00:00:00Z","docker_version":"1.9.0","os":"linux","architecture":"amd64","config":{"Labels":{"test":"schema1"}}}`
	body, err := json.Marshal(schema1.Manifest{
		Versioned:    schema1.ManifestSchemaVersion,
		Name:         "test/image",
		Tag:          "old",
		Architecture: "amd64",
		FSLayers: []schema1.FSLayer{
			{BlobSum: topLayer},
			{BlobSum: baseLayer},
		},
		History: []schema1.History{
			{V1Compatibility: config},
			{V1Compatibility: config},
		},
	})
	require.NoError(t, err)
	m, err := regmanifest.New(regmanifest.WithRaw(body))
	require.NoError(t, err)
	var result Manifest
	require.NoError(t, inspectSchema1(m, &result))
	assert.Equal(t, "old", result.Tag)
	assert.Equal(t, "1.9.0", result.DockerVersion)
	assert.Equal(t, "linux/amd64", result.Platform)
	assert.Equal(t, []string{baseLayer.String(), topLayer.String()}, result.Layers)
	assert.Equal(t, "schema1", result.Labels["test"])
	assert.NotNil(t, result.Created)
}

func TestManifest(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, ocispecs.Platform{
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
	assert.Equal(t, mediatype.Docker2Manifest, manifest.MIMEType)
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
	amd64Platform := ocispecs.Platform{Architecture: "amd64", OS: "linux"}
	armPlatform := ocispecs.Platform{Architecture: "arm", OS: "linux", Variant: "v7"}

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
	dbListWithoutPlatform := newTestManifestList(t,
		testManifestListInstance{manifest: oldARM.manifest, platform: armPlatform},
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
		Platform:      regplatform.Platform{OS: "linux", Architecture: "amd64"},
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
		{
			name:    "platform added to index",
			raw:     dbListWithoutPlatform.body,
			digest:  dbListWithoutPlatform.digest,
			updated: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manifest, updated, err := client.Manifest(image, Manifest{
				MIMEType: mediatype.Docker2ManifestList,
				Digest:   tt.digest,
				Raw:      tt.raw,
			})
			require.NoError(t, err)

			assert.Equal(t, tt.updated, updated)
			assert.Equal(t, registry.host()+"/acme/diun", manifest.Name)
			assert.Equal(t, "multi", manifest.Tag)
			assert.Equal(t, mediatype.Docker2ManifestList, manifest.MIMEType)
			assert.Equal(t, remoteList.digest, manifest.Digest)
			assert.Equal(t, "linux/amd64", manifest.Platform)
		})
	}
}

func TestManifestVariant(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	amd64Platform := ocispecs.Platform{Architecture: "amd64", OS: "linux"}
	armPlatform := ocispecs.Platform{Architecture: "arm", OS: "linux", Variant: "v7"}

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
		Platform: regplatform.Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
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

	_, ok := stderrors.AsType[NonImageArtifactError](err)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), ocispecs.MediaTypeEmptyJSON)
}

func TestManifestTaggedDigestUnknownTag(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	remoteImage := newTestRegistryImage(t, ocispecs.Platform{
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
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestManifestListCacheDoesNotLeakPlatformMetadata(t *testing.T) {
	r := newTestRegistry(t, "acme/diun")
	amd64 := ocispecs.Platform{OS: "linux", Architecture: "amd64"}
	arm := ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	amdImage := newTestRegistryImage(t, amd64, "25.0.0", map[string]string{"platform": "amd64"})
	armImage := newTestRegistryImage(t, arm, "25.0.0", map[string]string{"platform": "arm"})
	list := newTestManifestList(t,
		testManifestListInstance{manifest: amdImage.manifest, platform: amd64},
		testManifestListInstance{manifest: armImage.manifest, platform: arm},
	)
	r.addImage("amd64", amdImage)
	r.addImage("arm", armImage)
	r.addManifest("multi", list)
	image, err := ParseImage(ParseImageOptions{Name: r.imageName("multi")})
	require.NoError(t, err)
	amdClient := newTestRegistryClient(t, Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "amd64"}})
	armClient := newTestRegistryClient(t, Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}})
	cached, _, err := amdClient.Manifest(image, Manifest{})
	require.NoError(t, err)
	other, updated, err := armClient.Manifest(image, cached)
	require.NoError(t, err)
	require.False(t, updated)
	require.Equal(t, "linux/arm/v7", other.Platform)
	require.Equal(t, "arm", other.Labels["platform"])
	require.Equal(t, "amd64", cached.Labels["platform"])
	before := r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/multi")
	again, updated, err := armClient.Manifest(image, other)
	require.NoError(t, err)
	require.False(t, updated)
	require.Equal(t, other, again)
	require.Equal(t, before, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/multi"))
}

func TestManifestListCacheUsesSelectedInstance(t *testing.T) {
	for _, tt := range []struct {
		name     string
		platform ocispecs.Platform
		opts     Options
	}{
		{
			name:     "default host platform",
			platform: ocispecs.Platform{OS: runtime.GOOS, Architecture: runtime.GOARCH},
			opts:     Options{CompareDigest: true},
		},
		{
			name:     "ARM without explicit variant",
			platform: ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			opts:     Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm"}},
		},
		{
			name:     "ARM64 without explicit variant",
			platform: ocispecs.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"},
			opts:     Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm64"}},
		},
		{
			name:     "ARM compatible fallback variant",
			platform: ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v6"},
			opts:     Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry(t, "acme/diun")
			remote := newTestRegistryImage(t, tt.platform, "25.0.0", nil)
			list := newTestManifestList(t, testManifestListInstance{manifest: remote.manifest, platform: tt.platform})
			r.addImage("platform", remote)
			r.addManifest("multi", list)
			image, err := ParseImage(ParseImageOptions{Name: r.imageName("multi")})
			require.NoError(t, err)
			opts := tt.opts
			opts.InsecureTLS = true
			client, err := New(opts)
			require.NoError(t, err)
			cached, _, err := client.Manifest(image, Manifest{})
			require.NoError(t, err)
			require.Equal(t, remote.manifest.digest, cached.InstanceDigest)
			before := r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/multi")
			again, updated, err := client.Manifest(image, cached)
			require.NoError(t, err)
			require.False(t, updated)
			require.Equal(t, cached, again)
			require.Equal(t, before, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/multi"))
			// Existing databases are upgraded lazily by refreshing their metadata.
			cached.InstanceDigest = ""
			refreshed, updated, err := client.Manifest(image, cached)
			require.NoError(t, err)
			require.False(t, updated)
			require.Equal(t, remote.manifest.digest, refreshed.InstanceDigest)
			require.Equal(t, 1, r.requestCount(http.MethodGet, "/v2/acme/diun/manifests/"+list.digest.String()))
			require.Equal(t, 2, r.requestCount(http.MethodGet, "/v2/acme/diun/blobs/"+remote.config.digest.String()))
		})
	}
}

func TestManifestListCacheSeparatesARMVariants(t *testing.T) {
	for _, nested := range []bool{false, true} {
		r := newTestRegistry(t, "acme/diun")
		v6 := ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v6"}
		v7 := ocispecs.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
		image6 := newTestRegistryImage(t, v6, "25.0.0", map[string]string{"variant": "v6"})
		image7 := newTestRegistryImage(t, v7, "25.0.0", map[string]string{"variant": "v7"})
		r.addImage("v6", image6)
		r.addImage("v7", image7)
		list := newTestManifestList(t,
			testManifestListInstance{manifest: image6.manifest, platform: v6},
			testManifestListInstance{manifest: image7.manifest, platform: v7},
		)
		if nested {
			r.addManifest("inner", list)
			list = newTestManifestList(t,
				testManifestListInstance{manifest: list, platform: v6},
				testManifestListInstance{manifest: list, platform: v7},
			)
		}
		r.addManifest("multi", list)
		image, err := ParseImage(ParseImageOptions{Name: r.imageName("multi")})
		require.NoError(t, err)
		client6 := newTestRegistryClient(t, Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm", Variant: "v6"}})
		client7 := newTestRegistryClient(t, Options{CompareDigest: true, Platform: regplatform.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}})
		cached, _, err := client6.Manifest(image, Manifest{})
		require.NoError(t, err)
		other, updated, err := client7.Manifest(image, cached)
		require.NoError(t, err)
		require.False(t, updated)
		require.Equal(t, "v6", cached.Labels["variant"])
		require.Equal(t, "v7", other.Labels["variant"], "nested=%t", nested)
	}
}
