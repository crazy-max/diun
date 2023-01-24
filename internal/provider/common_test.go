package provider_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestValidateImage(t *testing.T) {
	cases := []struct {
		name          string
		image         string
		metadata      map[string]string
		labels        map[string]string
		watchByDef    bool
		imageDefaults model.Image
		expectedImage model.Image
		expectedErr   error
	}{

		{
			name:          "All excluded by default",
			image:         "myimg",
			expectedImage: model.Image{},
			expectedErr:   nil,
		},
		{
			name:       "Include using watch by default",
			image:      "myimg",
			watchByDef: true,
			expectedImage: model.Image{
				Name: "myimg",
			},
			expectedErr: nil,
		},
		{
			name:       "Include using global settings",
			image:      "myimg",
			watchByDef: true,
			imageDefaults: model.Image{
				WatchRepo: true,
				SortTags:  registry.SortTagSemver,
			},
			expectedImage: model.Image{
				Name:      "myimg",
				WatchRepo: true,
				SortTags:  registry.SortTagSemver,
			},
			expectedErr: nil,
		},
		{
			name:       "Override default image values with labels",
			image:      "myimg",
			watchByDef: true,
			labels: map[string]string{
				"diun.watch_repo": "false",
			},
			imageDefaults: model.Image{
				WatchRepo: true,
				SortTags:  registry.SortTagSemver,
			},
			expectedImage: model.Image{
				Name:      "myimg",
				WatchRepo: false,
				SortTags:  registry.SortTagSemver,
			},
			expectedErr: nil,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actualImg, actualErr := provider.ValidateImage(
				c.image,
				c.metadata,
				c.labels,
				c.watchByDef,
				c.imageDefaults,
			)
			assert.Equal(t, c.expectedImage, actualImg)
			assert.Equal(t, c.expectedErr, actualErr)
		})
	}
}
