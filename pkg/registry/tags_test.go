package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	assert.NotNil(t, rc)
	t.Parallel()

	cases := []struct {
		name                string
		imageName           string
		excludeOldVersions  bool
		expectedNotIncluded bool
		expectedExcluded    bool
		expectedContains    string
	}{
		{
			"parse image and tag",
			"crazymax/diun:3.0.0",
			false,
			false,
			false,
			"4.0.0",
		},
		{
			"parse image digest",
			"crazymax/diun:latest@sha256:3fca3dd86c2710586208b0f92d1ec4ce25382f4cad4ae76a2275db8e8bb24031",
			false,
			false,
			false,
			"4.0.0",
		},
		{
			"exclude older semver",
			"crazymax/diun:4.0.0",
			true,
			false,
			true,
			"4.20",
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			image, err := ParseImage(ParseImageOptions{
				Name: c.imageName,
			})
			if err != nil {
				t.Fatal(err)
			}

			tags, err := rc.Tags(TagsOptions{
				Image:              image,
				Sort:               SortTagSemver,
				ExcludeOldVersions: c.excludeOldVersions,
			})
			if err != nil {
				t.Fatal(err)
			}

			assert.Greater(t, tags.Total, 0)
			assert.Greater(t, len(tags.List), 0)
			// Make sure final list includes original tag and additional expected
			assert.Contains(t, tags.List, image.Tag)
			assert.Contains(t, tags.List, c.expectedContains)

			assert.Equal(t, c.expectedExcluded, tags.Excluded > 0, "Unexpected excluded tags")
			assert.Equal(t, c.expectedNotIncluded, tags.NotIncluded > 0, "Unexpected not included tags")
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
