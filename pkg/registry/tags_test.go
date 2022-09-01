package registry_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	assert.NotNil(t, rc)

	image, err := registry.ParseImage(registry.ParseImageOptions{
		Name: "crazymax/diun:3.0.0",
	})
	if err != nil {
		t.Error(err)
	}

	tags, err := rc.Tags(registry.TagsOptions{
		Image: image,
	})
	if err != nil {
		t.Error(err)
	}

	assert.True(t, tags.Total > 0)
	assert.True(t, len(tags.List) > 0)
}

func TestTagsExtraction(t *testing.T) {
	repotags := []string{
		"latest",
		"v-1.0.0",
		"version-1.1.0",
		"v-1.2.0",
		"version-1.2.3",
	}

	tags := registry.ExtractVersions(repotags, []string{`(v)-(\d+\.\d+\.\d+)`, `version-(\d+\.\d+\.\d+)`})
	assert.Equal(t, []string{"latest", "v1.0.0", "1.1.0", "v1.2.0", "1.2.3"}, tags)
}

func TestTagsSort(t *testing.T) {
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
		"edge",
		"latest",
	}

	testCases := []struct {
		name     string
		sortTag  registry.SortTag
		expected []string
	}{
		{
			name:    "sort default",
			sortTag: registry.SortTagDefault,
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
				"edge",
				"latest",
			},
		},
		{
			name:    "sort lexicographical",
			sortTag: registry.SortTagLexicographical,
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
				"edge",
				"latest",
			},
		},
		{
			name:    "sort reverse",
			sortTag: registry.SortTagReverse,
			expected: []string{
				"latest",
				"edge",
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
			sortTag: registry.SortTagSemver,
			expected: []string{
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tags := registry.SortTags(repotags, tt.sortTag)
			assert.Equal(t, tt.expected, tags)
		})
	}
}
