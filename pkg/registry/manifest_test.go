package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareDigest(t *testing.T) {
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
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	// download manifest
	_, _, err = rc.Manifest(img, Manifest{})
	assert.NoError(t, err)

	// check manifest
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

	// download manifest
	_, _, err = rc.Manifest(img, Manifest{})
	assert.NoError(t, err)

	// check manifest
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

	// download manifest
	_, _, err = rc.Manifest(img, Manifest{})
	assert.NoError(t, err)

	// check manifest
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

	// download manifest
	_, _, err = rc.Manifest(img, Manifest{})
	assert.NoError(t, err)

	// check manifest
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

func TestManifestTaggedDigest(t *testing.T) {
	rc, err := New(Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:4.25.0@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	})
	if err != nil {
		t.Error(err)
	}

	// download manifest
	_, _, err = rc.Manifest(img, Manifest{})
	assert.NoError(t, err)

	// check manifest
	manifest, updated, err := rc.Manifest(img, manifestCrazymaxDiun4250)
	assert.NoError(t, err)
	assert.Equal(t, false, updated)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "4.25.0", manifest.Tag)
	assert.Equal(t, "application/vnd.oci.image.index.v1+json", manifest.MIMEType)
	assert.Equal(t, "sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031", manifest.Digest.String())
	assert.Equal(t, "linux/amd64", manifest.Platform)
}

func TestManifestTaggedDigestUnknownTag(t *testing.T) {
	rc, err := New(Options{
		CompareDigest: true,
		ImageOs:       "linux",
		ImageArch:     "amd64",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:foo@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	})
	if err != nil {
		t.Error(err)
	}

	_, _, err = rc.Manifest(img, Manifest{})
	assert.Error(t, err)
}

var manifestCrazymaxDiun4250 = Manifest{
	Name:     "docker.io/crazymax/diun",
	Tag:      "4.25.0",
	MIMEType: "application/vnd.oci.image.index.v1+json",
	Digest:   "sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	Platform: "linux/amd64",
	Raw: []byte(`{
	"schemaVersion": 2,
	"mediaType": "application/vnd.oci.image.index.v1+json",
	"digest": "sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	"size": 4661,
	"manifests": [
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:bf782d6b2030c2a4c6884abb603ec5c99b5394554f57d56972cea24fb5d545d5",
			"size": 866,
			"platform": {
				"architecture": "386",
				"os": "linux"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:f44444abd33ee7c088d7527af84e3321f08313d12d9c679327bb8ae228e35f6a",
			"size": 866,
			"platform": {
				"architecture": "amd64",
				"os": "linux"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:df77b6ef88fbdb6175a2c60a9487a235aa1bdb39f60ee0a277d480d3cbc9f34a",
			"size": 866,
			"platform": {
				"architecture": "arm",
				"os": "linux",
				"variant": "v6"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:73e210387511588b38d16046de4ade809404b746cf6d16cd51ca23a96c8264b7",
			"size": 866,
			"platform": {
				"architecture": "arm",
				"os": "linux",
				"variant": "v7"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:1e070a6b2a3b5bf7c2c296fba6b01c8896514ae62aae6e48f4c28a775e5218dd",
			"size": 866,
			"platform": {
				"architecture": "arm64",
				"os": "linux"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:b7f984a85faf86839928fef6854f21da7afd2f2405b6043bf2aca562f1e1aa77",
			"size": 866,
			"platform": {
				"architecture": "ppc64le",
				"os": "linux"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:baa9a5e6de3f155526071eb0e55dcf14c12dca5c4301475e038df88fa5cb7c5a",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:bf782d6b2030c2a4c6884abb603ec5c99b5394554f57d56972cea24fb5d545d5",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:422bcf3cad62b4d8b21593387759889bcef02c28d7b0a3f6866b98b6502e8f01",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:f44444abd33ee7c088d7527af84e3321f08313d12d9c679327bb8ae228e35f6a",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:8ca5e335824bf17c10143c88f0e6955b5571dd69e06cd1a0ba46681169aa355d",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:df77b6ef88fbdb6175a2c60a9487a235aa1bdb39f60ee0a277d480d3cbc9f34a",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:01fdd0609476fe4da74af6bcb5a4fff97b0f9efbbea6b6ab142371ecc0738ffd",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:73e210387511588b38d16046de4ade809404b746cf6d16cd51ca23a96c8264b7",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:93178a24195f522195951a2cf16719bbae5358686b3789339c1096a85375117c",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:1e070a6b2a3b5bf7c2c296fba6b01c8896514ae62aae6e48f4c28a775e5218dd",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		},
		{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:1f5e5456e6f236c03684fea8070ca4095092a1d07a186acb03b15d160d100043",
			"size": 568,
			"annotations": {
				"vnd.docker.reference.digest": "sha256:b7f984a85faf86839928fef6854f21da7afd2f2405b6043bf2aca562f1e1aa77",
				"vnd.docker.reference.type": "attestation-manifest"
			},
			"platform": {
				"architecture": "unknown",
				"os": "unknown"
			}
		}
	]
}`)}
