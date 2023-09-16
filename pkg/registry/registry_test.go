package registry

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rc *Client
)

func TestMain(m *testing.M) {
	var err error

	rc, err = New(Options{
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
