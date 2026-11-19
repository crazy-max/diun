package registry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/regclient/regclient/types/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsPaginationOrigin(t *testing.T) {
	for _, tt := range []struct {
		name      string
		link      func(string, string) string
		wantError bool
	}{
		{
			name: "same origin",
			link: func(origin, _ string) string { return "<" + origin + `/v2/acme/diun/tags/list?page=2>; rel="next"` },
		},
		{
			name: "relative",
			link: func(_, _ string) string { return `<?page=2>; rel="next"` },
		},
		{
			name:      "cross origin",
			link:      func(_, target string) string { return "<" + target + `/private>; rel="next"` },
			wantError: true,
		},
		{
			name: "scheme relative cross origin",
			link: func(_, target string) string {
				return "<" + strings.TrimPrefix(target, "http:") + `/private>; rel="next"`
			},
			wantError: true,
		},
		{
			name: "plaintext downgrade",
			link: func(origin, _ string) string {
				return "<" + strings.Replace(origin, "https:", "http:", 1) + `/v2/acme/diun/tags/list?page=2>; rel="next"`
			},
			wantError: true,
		},
		{
			name: "second link cross origin",
			link: func(_, target string) string {
				return `<?page=2>; rel="prev"; title="one,two", <` + target + `/private>; rel="next"`
			},
			wantError: true,
		},
		{
			name:      "unbracketed link",
			link:      func(_, target string) string { return target + `/private; rel="next"` },
			wantError: true,
		},
		{
			name: "userinfo",
			link: func(origin, _ string) string {
				return "<" + strings.Replace(origin, "https://", "https://user:secret@", 1) + `/v2/acme/diun/tags/list?page=2>; rel="next"`
			},
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int32
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				_, _ = w.Write([]byte(`{"name":"acme/diun","tags":["leaked"]}`))
			}))
			t.Cleanup(target.Close)
			r := newTestRegistry(t, "acme/diun")
			r.addTagsPage("", []string{"first"}, tt.link(r.server.URL, target.URL))
			r.addTagsPage("2", []string{"second"}, "")
			image, err := ParseImage(ParseImageOptions{Name: r.imageName("latest")})
			require.NoError(t, err)
			client := newTestRegistryClient(t, Options{Timeout: time.Second})
			tags, err := client.ListTags(image)
			if tt.wantError {
				require.ErrorIs(t, err, errs.ErrNotRetryable)
				require.Nil(t, tags)
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"first", "second"}, tags)
			}
			assert.Zero(t, requests.Load(), "pagination must not contact another origin")
			if !tt.wantError {
				assert.Equal(t, 1, r.requestCount(http.MethodGet, "/v2/acme/diun/tags/list"))
			}
		})
	}
}

func TestTagsSelectPreservesSharedInput(t *testing.T) {
	tags := []string{"1.0.0", "3.0.0", "2.0.0"}
	reversed, err := (TagsOptions{Sort: SortTagReverse, Max: 2}).Select(tags)
	require.NoError(t, err)
	require.Equal(t, []string{"2.0.0", "3.0.0"}, reversed.List)
	require.Equal(t, []string{"1.0.0", "3.0.0", "2.0.0"}, tags)

	unchanged, err := (TagsOptions{Sort: SortTagDefault, Max: 1}).Select(tags)
	require.NoError(t, err)
	require.Equal(t, []string{"1.0.0"}, unchanged.List)
}

func TestTags(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	registry.addTagsPage("", []string{"latest", "1.0.0", "1.1.0", "1.1.1-beta"}, `</v2/acme/diun/tags/list?page=2>; rel="next"`)
	registry.addTagsPage("2", []string{"dev", "2.0.0", "nightly"}, "")

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.0.0"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{})
	tags, err := client.Tags(TagsOptions{
		Image:   image,
		Max:     2,
		Sort:    SortTagSemver,
		Include: []string{`^\d+\.\d+\.\d+$`, `^latest$`},
		Exclude: []string{`^1\.0\.0$`},
	})
	require.NoError(t, err)

	assert.Equal(t, &Tags{
		List:        []string{"2.0.0", "1.1.0"},
		NotIncluded: 3,
		Excluded:    1,
		Total:       7,
	}, tags)
}

func TestTagsSemverShorthandWithMax(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	registry.addTagsPage("", []string{
		"17.0-alpine", "17.1-alpine", "17.10-alpine", "17.11-alpine",
		"17.2-alpine", "17.9-alpine", "17.12", "latest",
	}, "")

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("17.10-alpine"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{})
	tags, err := client.Tags(TagsOptions{
		Image:   image,
		Max:     2,
		Sort:    SortTagSemver,
		Include: []string{`^17\.\d+-alpine$`},
	})
	require.NoError(t, err)

	assert.Equal(t, &Tags{
		List:        []string{"17.11-alpine", "17.10-alpine"},
		NotIncluded: 2,
		Total:       8,
	}, tags)
}

func TestTagsWithDigest(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t, "acme/diun")
	registry.addTagsPage("", []string{"latest"}, "")

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("latest") + "@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{})
	tags, err := client.Tags(TagsOptions{
		Image: image,
	})
	require.NoError(t, err)

	assert.Equal(t, &Tags{
		List:  []string{"latest"},
		Total: 1,
	}, tags)
}

func TestTagsSkipsGeneratedArtifactTags(t *testing.T) {
	registry := newTestRegistry(t, "acme/diun")
	registry.addTagsPage("", []string{
		"1.0.0",
		"sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526",
		"sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526.att",
		"sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526.sbom",
		"sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526.sig",
		"sha256-not-a-digest",
	}, "")

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.0.0"),
	})
	require.NoError(t, err)

	client := newTestRegistryClient(t, Options{})
	tags, err := client.Tags(TagsOptions{
		Image: image,
	})
	require.NoError(t, err)

	assert.Equal(t, &Tags{
		List:      []string{"1.0.0", "sha256-not-a-digest"},
		Artifacts: 4,
		Total:     6,
	}, tags)
}

func TestTagsWatchMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    WatchMode
		current string
		want    []string
		wantErr string
	}{
		{
			name:    "stable releases",
			mode:    WatchModeSemver,
			current: "1.5.0",
			want:    []string{"v1.5.1", "1.6.0", "2.0.0"},
		},
		{
			name:    "prereleases included",
			mode:    WatchModeSemverPrereleases,
			current: "v1.5.0",
			want:    []string{"v1.5.1", "1.6.0", "2.0.0-rc.1", "2.0.0", "1.6.0-alpine"},
		},
		{
			name:    "prerelease baseline",
			mode:    WatchModeSemverPrereleases,
			current: "2.0.0-beta.1",
			want:    []string{"2.0.0-rc.1", "2.0.0"},
		},
		{
			name:    "stable after prerelease",
			mode:    WatchModeSemver,
			current: "2.0.0-beta.1",
			want:    []string{"2.0.0"},
		},
		{
			name:    "nothing newer",
			mode:    WatchModeSemver,
			current: "3.0.0",
		},
		{
			name:    "configured tag only",
			mode:    WatchModeTag,
			current: "latest",
		},
		{
			name:    "invalid baseline",
			mode:    WatchModeSemver,
			current: "latest",
			wantErr: "requires a semantic version tag",
		},
		{
			name:    "shorthand baseline",
			mode:    WatchModeSemver,
			current: "1.5",
			wantErr: "requires a semantic version tag",
		},
		{
			name:    "date baseline",
			mode:    WatchModeSemverPrereleases,
			current: "20260616",
			wantErr: "requires a semantic version tag",
		},
		{
			name:    "invalid mode",
			mode:    "invalid",
			current: "1.5.0",
			wantErr: "unknown watch mode",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry(t, "acme/myapp")
			allTags := []string{
				"latest", "edge", "1.4.0", "1.5.0", "v1.5.0", "v1.5.1",
				"1.6.0", "2.0.0-rc.1", "2.0.0", "20260616", "nightly-20260616",
				"ubuntu-5.0", "1.6", "3", "1.6.0-alpine",
			}
			r.addTagsPage("", allTags, "")
			image, err := ParseImage(ParseImageOptions{Name: r.imageName(tt.current)})
			require.NoError(t, err)
			tags, err := newTestRegistryClient(t, Options{}).Tags(TagsOptions{
				Image:     image,
				WatchMode: tt.mode,
			})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Zero(t, r.requestCount("GET", "/v2/acme/myapp/tags/list"))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tags.List)
			if tt.mode == WatchModeTag {
				assert.Zero(t, tags.Total)
				assert.Zero(t, r.requestCount("GET", "/v2/acme/myapp/tags/list"))
			} else {
				assert.Equal(t, len(allTags), tags.Total)
				assert.Equal(t, len(allTags)-len(tt.want), tags.OutsideWatchMode)
			}
		})
	}
}

func TestTagsWatchModeFiltersBeforeMax(t *testing.T) {
	r := newTestRegistry(t, "acme/myapp")
	r.addTagsPage("", []string{
		"latest", "1.4.0", "1.5.0", "1.5.1", "1.6.0", "2.0.0-rc.1", "2.0.0",
		"sha256-64677ff7a877079df86d4a12e80e67a9548ea0facb2acb8c6719e79088e64526.sig",
	}, "")
	image, err := ParseImage(ParseImageOptions{Name: r.imageName("1.5.0")})
	require.NoError(t, err)
	tags, err := newTestRegistryClient(t, Options{}).Tags(TagsOptions{
		Image:     image,
		WatchMode: WatchModeSemver,
		Sort:      SortTagSemver,
		Max:       1,
		Include:   []string{`^\d`},
		Exclude:   []string{`^2\.0\.0$`},
	})
	require.NoError(t, err)
	assert.Equal(t, &Tags{
		List:             []string{"1.6.0"},
		NotIncluded:      1,
		Excluded:         1,
		Artifacts:        1,
		OutsideWatchMode: 3,
		Total:            8,
	}, tags)
}

func TestTagsWatchModeAll(t *testing.T) {
	r := newTestRegistry(t, "acme/myapp")
	allTags := []string{"latest", "1.0.0", "edge", "2.0.0-rc.1"}
	r.addTagsPage("", allTags, "")
	image, err := ParseImage(ParseImageOptions{Name: r.imageName("latest")})
	require.NoError(t, err)
	tags, err := newTestRegistryClient(t, Options{}).Tags(TagsOptions{
		Image:     image,
		WatchMode: WatchModeAll,
	})
	require.NoError(t, err)
	assert.Equal(t, allTags, tags.List)
	assert.Zero(t, tags.OutsideWatchMode)
}

func TestTagsWatchModeDigestOnly(t *testing.T) {
	image, err := ParseImage(ParseImageOptions{
		Name: "example.com/acme/myapp@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
	})
	require.NoError(t, err)
	_, err = newTestRegistryClient(t, Options{}).Tags(TagsOptions{
		Image:     image,
		WatchMode: WatchModeSemver,
	})
	require.ErrorContains(t, err, "requires a semantic version tag")
}

func TestTagsSortSemverShorthandWithSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "major minor",
			input:    []string{"17.0-alpine", "17.1-alpine", "17.10-alpine", "17.11-alpine", "17.2-alpine", "17.9-alpine"},
			expected: []string{"17.11-alpine", "17.10-alpine", "17.9-alpine", "17.2-alpine", "17.1-alpine", "17.0-alpine"},
		},
		{
			name:     "major only",
			input:    []string{"7-alpine", "8-alpine", "10-alpine", "9-alpine"},
			expected: []string{"10-alpine", "9-alpine", "8-alpine", "7-alpine"},
		},
		{
			name:     "mixed shorthand and full versions",
			input:    []string{"17-alpine", "17.1-alpine", "17.1.0-alpine", "17.1.1-alpine", "17.2-alpine", "17.10-alpine"},
			expected: []string{"17.10-alpine", "17.2-alpine", "17.1.1-alpine", "17.1.0-alpine", "17.1-alpine", "17-alpine"},
		},
		{
			name:     "release before prerelease",
			input:    []string{"17.11-alpine", "17.11", "17.10-alpine", "17.10"},
			expected: []string{"17.11", "17.11-alpine", "17.10", "17.10-alpine"},
		},
		{
			name:     "prefixes and dotted suffixes",
			input:    []string{"v17.9-alpine3.22", "postgres-17.11-alpine3.22", "17.10-alpine3.22"},
			expected: []string{"postgres-17.11-alpine3.22", "17.10-alpine3.22", "v17.9-alpine3.22"},
		},
		{
			name:     "build suffixes",
			input:    []string{"17.9+build.1", "17.11+build.1", "17.10+build.1"},
			expected: []string{"17.11+build.1", "17.10+build.1", "17.9+build.1"},
		},
		{
			name:     "invalid versions retain fallback order",
			input:    []string{"latest", "2025.04-alpine", "17.10-alpine", "17.01-alpine", "17.11-", "17.11.0.1-alpine", "edge"},
			expected: []string{"17.10-alpine", "17.11.0.1-alpine", "17.01-alpine", "17.11-", "2025.04-alpine", "edge", "latest"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, SortTags(tt.input, SortTagSemver))
		})
	}
}

func TestTagsSort(t *testing.T) {
	testCases := []struct {
		name     string
		sortTag  SortTag
		expected []string
	}{
		{
			name:    "sort default",
			sortTag: SortTagDefault,
			expected: []string{
				"0.1.0",
				"0.4.0",
				"3.0.0-beta.1",
				"3.0.0-beta.3",
				"3.0.0-beta.4",
				"4",
				"4.0.0",
				"4.0.0-beta.1",
				"4.1.0",
				"4.1.1",
				"4.10.0",
				"4.11.0",
				"4.12.0",
				"4.13.0",
				"4.14.0",
				"4.19.0",
				"4.2.0",
				"4.20",
				"4.20.0",
				"4.20.1",
				"4.21",
				"4.21.0",
				"4.3.0",
				"4.3.1",
				"4.4.0",
				"4.6.1",
				"4.7.0",
				"4.8.0",
				"4.8.1",
				"4.9.0",
				"ubuntu-5.0",
				"alpine-5.0",
				"edge",
				"latest",
			},
		},
		{
			name:    "sort lexicographical",
			sortTag: SortTagLexicographical,
			expected: []string{
				"0.1.0",
				"0.4.0",
				"3.0.0-beta.1",
				"3.0.0-beta.3",
				"3.0.0-beta.4",
				"4",
				"4.0.0",
				"4.0.0-beta.1",
				"4.1.0",
				"4.1.1",
				"4.10.0",
				"4.11.0",
				"4.12.0",
				"4.13.0",
				"4.14.0",
				"4.19.0",
				"4.2.0",
				"4.20",
				"4.20.0",
				"4.20.1",
				"4.21",
				"4.21.0",
				"4.3.0",
				"4.3.1",
				"4.4.0",
				"4.6.1",
				"4.7.0",
				"4.8.0",
				"4.8.1",
				"4.9.0",
				"alpine-5.0",
				"edge",
				"latest",
				"ubuntu-5.0",
			},
		},
		{
			name:    "sort reverse",
			sortTag: SortTagReverse,
			expected: []string{
				"latest",
				"edge",
				"alpine-5.0",
				"ubuntu-5.0",
				"4.9.0",
				"4.8.1",
				"4.8.0",
				"4.7.0",
				"4.6.1",
				"4.4.0",
				"4.3.1",
				"4.3.0",
				"4.21.0",
				"4.21",
				"4.20.1",
				"4.20.0",
				"4.20",
				"4.2.0",
				"4.19.0",
				"4.14.0",
				"4.13.0",
				"4.12.0",
				"4.11.0",
				"4.10.0",
				"4.1.1",
				"4.1.0",
				"4.0.0-beta.1",
				"4.0.0",
				"4",
				"3.0.0-beta.4",
				"3.0.0-beta.3",
				"3.0.0-beta.1",
				"0.4.0",
				"0.1.0",
			},
		},
		{
			name:    "sort semver",
			sortTag: SortTagSemver,
			expected: []string{
				"alpine-5.0",
				"ubuntu-5.0",
				"4.21.0",
				"4.21",
				"4.20.1",
				"4.20.0",
				"4.20",
				"4.19.0",
				"4.14.0",
				"4.13.0",
				"4.12.0",
				"4.11.0",
				"4.10.0",
				"4.9.0",
				"4.8.1",
				"4.8.0",
				"4.7.0",
				"4.6.1",
				"4.4.0",
				"4.3.1",
				"4.3.0",
				"4.2.0",
				"4.1.1",
				"4.1.0",
				"4.0.0",
				"4",
				"4.0.0-beta.1",
				"3.0.0-beta.4",
				"3.0.0-beta.3",
				"3.0.0-beta.1",
				"0.4.0",
				"0.1.0",
				"edge",
				"latest",
			},
		},
	}

	for _, tt := range testCases {
		repotags := []string{
			"0.1.0",
			"0.4.0",
			"3.0.0-beta.1",
			"3.0.0-beta.3",
			"3.0.0-beta.4",
			"4",
			"4.0.0",
			"4.0.0-beta.1",
			"4.1.0",
			"4.1.1",
			"4.10.0",
			"4.11.0",
			"4.12.0",
			"4.13.0",
			"4.14.0",
			"4.19.0",
			"4.2.0",
			"4.20",
			"4.20.0",
			"4.20.1",
			"4.21",
			"4.21.0",
			"4.3.0",
			"4.3.1",
			"4.4.0",
			"4.6.1",
			"4.7.0",
			"4.8.0",
			"4.8.1",
			"4.9.0",
			"ubuntu-5.0",
			"alpine-5.0",
			"edge",
			"latest",
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tags := SortTags(repotags, tt.sortTag)
			assert.Equal(t, tt.expected, tags)
		})
	}
}
