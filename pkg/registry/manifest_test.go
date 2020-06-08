package registry_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestManifestVariant(t *testing.T) {
	rc, err := registry.New(registry.Options{
		ImageOs:      "linux",
		ImageArch:    "arm",
		ImageVariant: "v7",
	})
	if err != nil {
		panic(err.Error())
	}

	img, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:2.5.0",
	})
	if err != nil {
		t.Error(err)
	}

	manifest, err := rc.Manifest(img)
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/crazymax/diun", manifest.Name)
	assert.Equal(t, "2.5.0", manifest.Tag)
	assert.Equal(t, "application/vnd.docker.distribution.manifest.list.v2+json", manifest.MIMEType)
	assert.Equal(t, "linux/arm/v7", manifest.Platform)
	assert.Empty(t, manifest.DockerVersion)
}
