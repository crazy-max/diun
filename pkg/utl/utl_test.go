package utl_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/stretchr/testify/assert"
)

func TestExtractCaptureRegex(t *testing.T) {
	assert.Equal(t, "version-1.2.3", utl.ExtractCaptureRegex("version-1.2.3", "^not-matching-regex$"))
	assert.Equal(t, "version-1.2.3", utl.ExtractCaptureRegex("version-1.2.3", `version-1\.2\.3`))
	assert.Equal(t, "v1.2.3", utl.ExtractCaptureRegex("version-1.2.3", `(v)ersion-(1\.2\.3)`))
	assert.Equal(t, "1.2.3", utl.ExtractCaptureRegex("version-1.2.3", `version-(1\.2\.3)`))
	assert.Equal(t, "1.2.3", utl.ExtractCaptureRegex("version-1.2.3", `^version-(1\.2\.3)$`))
}
