package registry_test

import (
	"os"
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

var (
	rc *registry.Client
)

func TestMain(m *testing.M) {
	var err error

	rc, err = registry.New(registry.Options{
		ImageOs:   "linux",
		ImageArch: "amd64",
	})
	if err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	assert.NotNil(t, rc)
}

func TestTags(t *testing.T) {
	assert.NotNil(t, rc)

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:3.0.0",
	})
	if err != nil {
		t.Error(err)
	}

	tags, err := rc.Tags(registry.TagsOptions{
		Image: image,
	})
	if err != nil {
		t.Error(err)
	}

	assert.True(t, tags.Total > 0)
	assert.True(t, len(tags.List) > 0)
}
