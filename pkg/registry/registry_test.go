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
