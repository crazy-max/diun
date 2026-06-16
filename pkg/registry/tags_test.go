package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNormalizeSemver(t *testing.T) {
	cases := []struct {
		tag      string
		expected string
	}{
		{"1.5.0", "v1.5.0"},
		{"v1.5.0", "v1.5.0"},
		{"1.5.0-rc.1", "v1.5.0-rc.1"},
		{"v2.0.0-alpha", "v2.0.0-alpha"},
		{"latest", ""},
		{"edge", ""},
		{"", ""},
		{"4", "v4"},
		{"ubuntu-5.0", "v5.0"},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			assert.Equal(t, tc.expected, normalizeSemver(tc.tag))
		})
	}
}

func TestTagsMinSemver(t *testing.T) {
	registry := newTestRegistry(t, "acme/myapp")
	registry.addTagsPage("", []string{
		"latest",
		"edge",
		"1.4.0",
		"1.5.0",
		"1.5.1",
		"1.6.0",
		"2.0.0",
		"2.0.0-rc.1",
		"2.0.0-beta",
	}, "")

	image, err := ParseImage(ParseImageOptions{
		Name: registry.imageName("1.5.0"),
	})
	require.NoError(t, err)
	client := newTestRegistryClient(t, Options{})

	t.Run("stable only, no prereleases", func(t *testing.T) {
		tags, err := client.Tags(TagsOptions{
			Image:              image,
			MinSemver:          "1.5.0",
			IncludePrereleases: false,
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"1.5.1", "1.6.0", "2.0.0"}, tags.List)
		// latest, edge, 1.4.0, 1.5.0, 2.0.0-rc.1, 2.0.0-beta all filtered
		assert.Equal(t, 6, tags.OlderOrEqual)
		assert.Equal(t, 9, tags.Total)
	})

	t.Run("include prereleases", func(t *testing.T) {
		tags, err := client.Tags(TagsOptions{
			Image:              image,
			MinSemver:          "1.5.0",
			IncludePrereleases: true,
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"1.5.1", "1.6.0", "2.0.0", "2.0.0-rc.1", "2.0.0-beta"}, tags.List)
		assert.Equal(t, 4, tags.OlderOrEqual) // latest, edge, 1.4.0, 1.5.0
		assert.Equal(t, 9, tags.Total)
	})

	t.Run("min semver with v prefix", func(t *testing.T) {
		tags, err := client.Tags(TagsOptions{
			Image:              image,
			MinSemver:          "v1.5.0",
			IncludePrereleases: false,
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"1.5.1", "1.6.0", "2.0.0"}, tags.List)
	})

	t.Run("non-semver min version disables filtering", func(t *testing.T) {
		tags, err := client.Tags(TagsOptions{
			Image:     image,
			MinSemver: "latest",
		})
		require.NoError(t, err)
		// "latest" can't be parsed as semver → no filtering applied
		assert.Equal(t, 9, len(tags.List))
		assert.Equal(t, 0, tags.OlderOrEqual)
	})
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
