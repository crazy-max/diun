package dockerfile_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/dockerfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromImages(t *testing.T) {
	c, err := dockerfile.New(dockerfile.Options{
		Filename: "./fixtures/valid.Dockerfile",
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	img, err := c.FromImages()
	require.NoError(t, err)
	require.NotNil(t, img)
	require.Equal(t, 3, len(img))

	assert.Equal(t, "alpine:3.14", img[0].Name)
	assert.Equal(t, 5, img[0].Line)
	assert.Equal(t, []string{"diun.platform=linux/amd64"}, img[0].Comments)

	assert.Equal(t, "crazymax/yasu", img[1].Name)
	assert.Equal(t, 10, img[1].Line)
	assert.Equal(t, []string{"diun.watch_repo=true", "diun.max_tags=10", "diun.platform=linux/amd64"}, img[1].Comments)

	assert.Equal(t, "crazymax/docker:20.10.6", img[2].Name)
	assert.Equal(t, 15, img[2].Line)
	assert.Equal(t, []string{"diun.watch_repo=true", "diun.include_tags=^\\d+\\.\\d+\\.\\d+$", "diun.platform=linux/amd64"}, img[2].Comments)
}
