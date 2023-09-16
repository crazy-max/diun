package dockerfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromImages(t *testing.T) {
	c, err := New(Options{
		Filename: "./fixtures/valid.Dockerfile",
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	img, err := c.FromImages()
	require.NoError(t, err)
	require.NotNil(t, img)
	require.Equal(t, 4, len(img))

	assert.Equal(t, "alpine:3.14", img[0].Name)
	assert.Equal(t, 5, img[0].Line)
	assert.Equal(t, []string{"diun.platform=linux/amd64"}, img[0].Comments)

	assert.Equal(t, "crazymax/yasu", img[1].Name)
	assert.Equal(t, 10, img[1].Line)
	assert.Equal(t, []string{"diun.watch_repo=true", "diun.max_tags=10", "diun.platform=linux/amd64"}, img[1].Comments)

	assert.Equal(t, "crazymax/docker:20.10.6", img[2].Name)
	assert.Equal(t, 15, img[2].Line)
	assert.Equal(t, []string{"diun.watch_repo=true", "diun.include_tags=^\\d+\\.\\d+\\.\\d+$", "diun.platform=linux/amd64"}, img[2].Comments)

	assert.Equal(t, "crazymax/ddns-route53:foo@sha256:9cb3af44cdd00615266c87e60bc05cac534297be14c4596800b57322f9313615", img[3].Name)
	assert.Equal(t, 21, img[3].Line)
	assert.Equal(t, []string{"diun.platform=linux/amd64", "diun.metadata.foo=bar"}, img[3].Comments)
}
