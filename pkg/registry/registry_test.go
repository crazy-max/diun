package registry_test

import (
	"os"
	"testing"

	"github.com/crazy-max/diun/pkg/registry"
	"github.com/stretchr/testify/assert"
)

var (
	rc *registry.Client
)

func TestMain(m *testing.M) {
	var err error

	rc, err = registry.New(registry.Options{})
	if err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	assert.NotNil(t, rc)
}
