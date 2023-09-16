package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareDigest(t *testing.T) {
	t.Parallel()
	rc, err := New(Options{
		CompareDigest: true,
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, _, err := rc.Manifest(img, Manifest{
		Name:     "docker.io/crazymax/diun",
		Tag:      "2.5.0",
		MIMEType: "application/vnd.docker.distribution.manifest.list.v2+json",
		Digest:   "sha256:db618981ef3d07699ff6cd8b9d2a81f51a021747bc08c85c1b0e8d11130c2be5",
		Platform: "linux/amd64",
	})
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "2.5.0", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "linux/amd64", manifest.Platform)
	assert.Empty(t, manifest.DockerVersion)
}

func TestManifest(t *testing.T) {
	t.Parallel()
	rc, err := New(Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "portainer/portainer-ce:linux-amd64-2.5.1",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, updated, err := rc.Manifest(img, Manifest{
		Name:     "docker.io/portainer/portainer-ce",
		Tag:      "linux-amd64-2.5.1",
		MIMEType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:   "sha256:653057af0d2d961f436c75deda1ca7fe3defc89664bed6bd3da8c91c88c1ce05",
		Platform: "linux/amd64",
		Raw: []byte(`{
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "schemaVersion": 2,
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "digest": "sha256:45be17a5903a1129362792537fc6b18bc91fe03e2581501b514ac5d45ede128e",
      "size": 1704
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "digest": "sha256:94cfa856b2b17d5e36c7df9875ebbbed7e939a8292df5fe22d2dfce0434330f2",
         "size": 122403
      },
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "digest": "sha256:49d59ee0881a4f04166d438b27055e2b29327abbbb0f274951255ee880912056",
         "size": 92
      },
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "digest": "sha256:3fc1bc38fb56bce4b3913d0fab6072822541142793a6d997a4a69d5d81fa46e0",
         "size": 74850629
      }
   ]
}`),
	})

	assert.NoError(t, err)
	assert.Equal(t, false, updated)
	assert.Equal(t, "docker.io/portainer/portainer-ce", manifest.Name)
	assert.Equal(t, "linux-amd64-2.5.1", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", manifest.MIMEType)
	assert.Equal(t, "sha256:653057af0d2d961f436c75deda1ca7fe3defc89664bed6bd3da8c91c88c1ce05", manifest.Digest.String())
	assert.Equal(t, "linux/amd64", manifest.Platform)
}

func TestManifestMultiUpdatedPlatform(t *testing.T) {
	t.Parallel()
	rc, err := New(Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "mongo:3.6.21",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, updated, err := rc.Manifest(img, Manifest{
		Name:     "docker.io/library/mongo",
		Tag:      "3.6.21",
		MIMEType: "application/vnd.docker.distribution.manifest.list.v2+json",
		Digest:   "sha256:61f5dce8422d36b2a4ad0077bc499b1b68320e13fd30aa0b201c080fef42a39a",
		Platform: "linux/amd64",
		Raw: []byte(`{
  "manifests": [
    {
      "digest": "sha256:98f22b0bf33479e2c34d99c820d9ded79cdf46b2c6f54af5a11191a90ff369ed",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      },
      "size": 3030
    },
    {
      "digest": "sha256:8226c9734c19533d5cc52748e35ae10085f3b4ef0a3bd4537017bc2484589511",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm64",
        "os": "linux",
        "variant": "v8"
      },
      "size": 3030
    },
    {
      "digest": "sha256:fb9e9376b228ba8d75d62b10aadaa3ed445266f85e27af3da531666d992f9621",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "windows",
        "os.version": "10.0.17763.1697"
      },
      "size": 2771
    },
    {
      "digest": "sha256:f0534dfb20d90f152a7b4ae8812c61381cff7de983c2b17fc1fe3558a237fdac",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "windows",
        "os.version": "10.0.14393.4169"
      },
      "size": 2693
    }
  ],
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "schemaVersion": 2
}`),
	})

	assert.NoError(t, err)
	assert.Equal(t, true, updated)
	assert.Equal(t, "docker.io/library/mongo", manifest.Name)
	assert.Equal(t, "3.6.21", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "sha256:3cff2069adb34a330552695659c261bca69148e325863763b78b0285dd1a25c9", manifest.Digest.String())
	assert.Equal(t, "linux/amd64", manifest.Platform)
}

func TestManifestMultiNotUpdatedPlatform(t *testing.T) {
	t.Parallel()
	rc, err := New(Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "mongo:3.6.21",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, updated, err := rc.Manifest(img, Manifest{
		Name:     "docker.io/library/mongo",
		Tag:      "3.6.21",
		MIMEType: "application/vnd.docker.distribution.manifest.list.v2+json",
		Digest:   "sha256:61f5dce8422d36b2a4ad0077bc499b1b68320e13fd30aa0b201c080fef42a39a",
		Platform: "linux/amd64",
		Raw: []byte(`{
  "manifests": [
    {
      "digest": "sha256:6e5d3405a510988d96f0fa3ec7220040be27ce783eb4cd576feb1a69b382ea20",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      },
      "size": 3030
    },
    {
      "digest": "sha256:8226c9734c19533d5cc52748e35ae10085f3b4ef0a3bd4537017bc2484589511",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm64",
        "os": "linux",
        "variant": "v8"
      },
      "size": 3030
    },
    {
      "digest": "sha256:0fcde35d138739e27b79a8b9863dedc1fdd65fd3a82a319842f86edc87d11594",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "windows",
        "os.version": "10.0.17763.1817"
      },
      "size": 2771
    },
    {
      "digest": "sha256:6f54fda6a88a56c0953e901f0285a74a16b4cf1bec021b2434e3bfe78cabfada",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "windows",
        "os.version": "10.0.14393.4283"
      },
      "size": 2693
    }
  ],
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "schemaVersion": 2
}`),
	})

	assert.NoError(t, err)
	assert.Equal(t, false, updated)
	assert.Equal(t, "docker.io/library/mongo", manifest.Name)
	assert.Equal(t, "3.6.21", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "sha256:3cff2069adb34a330552695659c261bca69148e325863763b78b0285dd1a25c9", manifest.Digest.String())
	assert.Equal(t, "linux/amd64", manifest.Platform)
}

func TestManifestVariant(t *testing.T) {
	t.Parallel()
	rc, err := New(Options{
		ImageOs:      "linux",
		ImageArch:    "arm",
		ImageVariant: "v7",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, _, err := rc.Manifest(img, Manifest{})
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "2.5.0", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "linux/arm/v7", manifest.Platform)
	assert.Empty(t, manifest.DockerVersion)
}
