package dockerfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListJobParsesDockerfileImages(t *testing.T) {
	dockerfile := filepath.Join(t.TempDir(), "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfile, []byte(`
# non-diun comment
# diun.max_tags=5
# diun.include_tags=^3\.
# diun.metadata.owner=ops
FROM alpine:3.19

FROM scratch
`), 0600))

	jobs := New(&model.PrdDockerfile{
		Patterns: []string{dockerfile},
	}, &model.Defaults{
		SortTags: registry.SortTagSemver,
	}).ListJob()

	require.Len(t, jobs, 1)
	assert.Equal(t, "dockerfile", jobs[0].Provider)
	assert.Equal(t, model.Image{
		Name:        "alpine:3.19",
		MaxTags:     5,
		SortTags:    registry.SortTagSemver,
		IncludeTags: []string{"^3\\."},
		Metadata: map[string]string{
			"owner": "ops",
		},
	}, jobs[0].Image)
}

func TestListJobReturnsEmptyWithoutConfig(t *testing.T) {
	assert.Empty(t, New(nil, nil).ListJob())
}
