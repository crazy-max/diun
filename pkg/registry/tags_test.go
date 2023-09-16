package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	assert.NotNil(t, rc)

	image, err := ParseImage(ParseImageOptions{
		Name: "crazymax/diun:3.0.0",
	})
	if err != nil {
		t.Error(err)
	}

	tags, err := rc.Tags(TagsOptions{
		Image: image,
	})
	if err != nil {
		t.Error(err)
	}

	assert.True(t, tags.Total > 0)
	assert.True(t, len(tags.List) > 0)
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
		tt := tt
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
