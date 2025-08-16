package model_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDedupeImageList(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []model.Image
		expected []model.Image
	}{
		{
			desc: "dedupe",
			input: []model.Image{
				{
					Name:        "alpine",
					IncludeTags: []string{"latest"},
				},
				{
					Name:        "alpine",
					IncludeTags: []string{"latest"},
				},
				{
					Name:        "alpine",
					IncludeTags: []string{"oldest"},
				},
			},
			expected: []model.Image{
				{
					Name:        "alpine",
					IncludeTags: []string{"latest"},
				},
				{
					Name:        "alpine",
					IncludeTags: []string{"oldest"},
				},
			},
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			result := model.ImageList(tt.input).Dedupe()
			assert.Equal(t, tt.expected, result)
		})
	}
}
