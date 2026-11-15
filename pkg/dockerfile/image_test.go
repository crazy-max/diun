package dockerfile

import (
	"os"
	"path/filepath"
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

func TestFromImagesMultiStage(t *testing.T) {
	c, err := New(Options{
		Filename: "./fixtures/multistage.Dockerfile",
	})
	require.NoError(t, err)

	img, err := c.FromImages()
	require.NoError(t, err)
	require.Len(t, img, 2)

	assert.Equal(t, "alpine:3.14", img[0].Name)
	assert.Equal(t, 1, img[0].Line)

	assert.Equal(t, "crazymax/yasu", img[1].Name)
	assert.Equal(t, 6, img[1].Line)
}

func TestFromImagesStageReferences(t *testing.T) {
	cases := []struct {
		name       string
		dockerfile string
		want       []string
		wantErr    string
	}{
		{
			name: "copy from first stage",
			dockerfile: `FROM alpine:3.14 AS base
FROM scratch AS final
COPY --from=0 /etc/os-release /os-release
`,
			want: []string{"alpine:3.14"},
		},
		{
			name: "copy from second stage",
			dockerfile: `FROM scratch AS first
FROM alpine:3.14 AS second
FROM scratch AS final
COPY --from=1 /etc/os-release /os-release
`,
			want: []string{"alpine:3.14"},
		},
		{
			name: "copy from mixed case stage",
			dockerfile: `FROM alpine:3.14 AS Base
FROM scratch AS final
COPY --from=bAsE /etc/os-release /os-release
`,
			want: []string{"alpine:3.14"},
		},
		{
			name: "mount from mixed case stage",
			dockerfile: `FROM alpine:3.14 AS Base
FROM alpine:3.14 AS final
RUN --mount=type=bind,from=bAsE,target=/base true
`,
			want: []string{"alpine:3.14", "alpine:3.14"},
		},
		{
			name: "numeric base image",
			dockerfile: `FROM 0 AS base
FROM scratch AS final
COPY --from=base /etc/os-release /os-release
`,
			want: []string{"0"},
		},
		{
			name: "numeric mount image",
			dockerfile: `FROM alpine:3.14
RUN --mount=type=bind,from=0,target=/base true
`,
			want: []string{"alpine:3.14", "0"},
		},
		{
			name: "external copy image",
			dockerfile: `FROM alpine:3.14 AS base
COPY --from=0:latest /etc/os-release /os-release
`,
			want: []string{"alpine:3.14", "0:latest"},
		},
		{
			name: "out of range stage index",
			dockerfile: `FROM alpine:3.14 AS base
COPY --from=1 /etc/os-release /os-release
`,
			wantErr: "invalid stage index 1",
		},
		{
			name: "negative stage index",
			dockerfile: `FROM alpine:3.14 AS base
COPY --from=-1 /etc/os-release /os-release
`,
			wantErr: "invalid stage index -1",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "Dockerfile")
			require.NoError(t, os.WriteFile(filename, []byte(tt.dockerfile), 0600))
			c, err := New(Options{
				Filename: filename,
			})
			require.NoError(t, err)

			images, err := c.FromImages()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, images, len(tt.want))
			for i, name := range tt.want {
				assert.Equal(t, name, images[i].Name)
			}
		})
	}
}
