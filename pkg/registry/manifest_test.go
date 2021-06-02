package registry_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestCompareDigest(t *testing.T) {
	rc, err := registry.New(registry.Options{
		CompareDigest: true,
	})
	if err != nil {
		t.Error(err)
	}

	img, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, err := rc.Manifest(img, registry.Manifest{
		Name:          "docker.io/crazymax/diun",
		Tag:           "2.5.0",
		MIMEType:      "application/vnd.docker.distribution.manifest.list.v2+json",
		Digest:        "sha256:db618981ef3d07699ff6cd8b9d2a81f51a021747bc08c85c1b0e8d11130c2be5",
		DockerVersion: "",
		Labels: map[string]string{
			"maintainer":                      "CrazyMax",
			"org.label-schema.build-date":     "2020-03-01T18:00:42Z",
			"org.label-schema.description":    "Docker image update notifier",
			"org.label-schema.name":           "Diun",
			"org.label-schema.schema-version": "1.0",
			"org.label-schema.url":            "https://github.com/crazy-max/diun",
			"org.label-schema.vcs-ref":        "488ce441",
			"org.label-schema.vcs-url":        "https://github.com/crazy-max/diun",
			"org.label-schema.vendor":         "CrazyMax",
			"org.label-schema.version":        "2.5.0",
		},
		Platform: "linux/amd64",
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "2.5.0", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "linux/amd64", manifest.Platform)
	assert.Empty(t, manifest.DockerVersion)
}

func TestManifestVariant(t *testing.T) {
	rc, err := registry.New(registry.Options{
		ImageOs:      "linux",
		ImageArch:    "arm",
		ImageVariant: "v7",
	})
	if err != nil {
		t.Error(err)
	}

	img, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, err := rc.Manifest(img, registry.Manifest{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "2.5.0", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "linux/arm/v7", manifest.Platform)
	assert.Empty(t, manifest.DockerVersion)
}
