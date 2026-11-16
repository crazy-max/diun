package model

import (
	"testing"

	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/require"
)

func TestResolveWatchMode(t *testing.T) {
	tests := []struct {
		name     string
		image    Image
		defaults *Defaults
		want     registry.WatchMode
		wantErr  bool
	}{
		{
			name: "default tag",
			want: registry.WatchModeTag,
		},
		{
			name:     "default mode",
			defaults: &Defaults{WatchMode: registry.WatchModeSemver},
			want:     registry.WatchModeSemver,
		},
		{
			name:     "legacy default true",
			defaults: &Defaults{WatchRepo: new(true)},
			want:     registry.WatchModeAll,
		},
		{
			name:     "legacy default false",
			defaults: &Defaults{WatchRepo: new(false)},
			want:     registry.WatchModeTag,
		},
		{
			name: "mode beats legacy default",
			defaults: &Defaults{
				WatchMode: registry.WatchModeSemver,
				WatchRepo: new(true),
			},
			want: registry.WatchModeSemver,
		},
		{
			name:     "image legacy false beats default mode",
			image:    Image{WatchRepo: new(false)},
			defaults: &Defaults{WatchMode: registry.WatchModeAll},
			want:     registry.WatchModeTag,
		},
		{
			name:     "image legacy true beats default mode",
			image:    Image{WatchRepo: new(true)},
			defaults: &Defaults{WatchMode: registry.WatchModeTag},
			want:     registry.WatchModeAll,
		},
		{
			name:     "image mode beats legacy default",
			image:    Image{WatchMode: registry.WatchModeTag},
			defaults: &Defaults{WatchRepo: new(true)},
			want:     registry.WatchModeTag,
		},
		{
			name: "image mode beats image legacy",
			image: Image{
				WatchMode: registry.WatchModeSemverPrereleases,
				WatchRepo: new(false),
			},
			defaults: &Defaults{WatchMode: registry.WatchModeAll},
			want:     registry.WatchModeSemverPrereleases,
		},
		{
			name:    "invalid image mode",
			image:   Image{WatchMode: "invalid"},
			wantErr: true,
		},
		{
			name:     "invalid default mode",
			defaults: &Defaults{WatchMode: "invalid"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.image.ResolveWatchMode(tt.defaults)
			if tt.wantErr {
				require.ErrorContains(t, err, "unknown watch mode")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, tt.image.WatchMode)
		})
	}
}
