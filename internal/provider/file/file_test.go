package file

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	defaults = model.Defaults{
		NotifyOn: model.NotifyOnDefaults,
		SortTags: registry.SortTagReverse,
		MaxTags:  25,
		IncludeTags: []string{
			`^(0|[1-9]\d*)\..*`,
		},
		ExcludeTags: []string{
			`^0\.0\..*`,
		},
	}
	bintrayFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
				WatchMode: registry.WatchModeTag,
				RegOpt:    "bintrayoptions",
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.bintray.io/jfrog/xray-server:2.8.6",
				WatchMode: registry.WatchModeAll,
				WatchRepo: new(true),
				NotifyOn: []model.NotifyOn{
					model.NotifyOnNew,
				},
				SortTags: registry.SortTagLexicographical,
				MaxTags:  50,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
	}
	dockerhubFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.io/crazymax/nextcloud:latest",
				WatchMode: registry.WatchModeTag,
				RegOpt:    "myregistry",
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/swarm-cronjob",
				WatchMode: registry.WatchModeAll,
				WatchRepo: new(true),
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagSemver,
				MaxTags:   25,
				IncludeTags: []string{
					`^1\.2\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.io/portainer/portainer",
				WatchMode: registry.WatchModeAll,
				WatchRepo: new(true),
				NotifyOn:  model.NotifyOnDefaults,
				MaxTags:   10,
				SortTags:  registry.SortTagReverse,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "traefik",
				WatchMode: registry.WatchModeAll,
				WatchRepo: new(true),
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagDefault,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
					`latest`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "alpine",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				Platform: model.ImagePlatform{
					OS:      "linux",
					Arch:    "arm64",
					Variant: "v8",
				},
				MaxTags: 25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.io/graylog/graylog:3.2.0",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "jacobalberty/unifi:5.9",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/ddns-route53",
				WatchMode: registry.WatchModeAll,
				WatchRepo: new(true),
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^1\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
	}
	quayFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "quay.io/coreos/hyperkube",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
	}
	lscrFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "lscr.io/linuxserver/heimdall",
				WatchMode: registry.WatchModeTag,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagReverse,
				HubLink:   "https://fleet.linuxserver.io/image?name=linuxserver/heimdall",
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`^0\.0\..*`,
				},
			},
		},
	}
)

func TestListJobFilename(t *testing.T) {
	fc := New(&model.PrdFile{
		Filename: "./fixtures/dockerhub.yml",
	}, &defaults)

	assert.Equal(t, dockerhubFile, fc.ListJob())
}

func TestListJobDirectory(t *testing.T) {
	fc := New(&model.PrdFile{
		Directory: "./fixtures",
	}, &defaults)

	assert.Equal(t, append(append(bintrayFile, dockerhubFile...), append(lscrFile, quayFile...)...), fc.ListJob())
}

func TestDefaultImageOptions(t *testing.T) {
	fc := New(&model.PrdFile{
		Filename: "./fixtures/dockerhub.yml",
	}, &model.Defaults{
		WatchRepo: new(true),
	})

	for _, job := range fc.ListJob() {
		assert.True(t, *job.Image.WatchRepo)
	}
}

func TestWatchRepoDeprecationWarning(t *testing.T) {
	tests := []struct {
		name   string
		fields string
		warn   bool
	}{
		{
			name: "inherited legacy default",
		},
		{
			name:   "modern option",
			fields: "  watch_mode: semver\n",
		},
		{
			name:   "legacy true",
			fields: "  watch_repo: true\n",
			warn:   true,
		},
		{
			name:   "legacy false",
			fields: "  watch_repo: false\n",
			warn:   true,
		},
		{
			name:   "both options",
			fields: "  watch_repo: true\n  watch_mode: semver\n",
			warn:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "images.yml")
			require.NoError(t, os.WriteFile(filename, []byte("- name: alpine:3.20.0\n"+tt.fields), 0o600))
			var output bytes.Buffer
			client := &Client{
				config:   &model.PrdFile{Filename: filename},
				defaults: &model.Defaults{WatchRepo: new(true)},
				logger:   zerolog.New(&output),
			}
			require.Len(t, client.listFileImage(), 1)
			assert.Equal(t, tt.warn, strings.Contains(output.String(), "is deprecated"))
			if tt.warn {
				assert.Contains(t, output.String(), `"level":"warn"`)
				assert.Contains(t, output.String(), "use watch_mode")
				assert.Contains(t, output.String(), `"img_name":"alpine:3.20.0"`)
			}
		})
	}
}

func TestListJobWatchMode(t *testing.T) {
	tests := []struct {
		name     string
		fields   string
		defaults model.Defaults
		want     registry.WatchMode
		wantErr  bool
	}{
		{
			name: "default mode",
			want: registry.WatchModeTag,
		},
		{
			name:   "explicit mode",
			fields: "  watch_mode: semver+prereleases\n",
			want:   registry.WatchModeSemverPrereleases,
		},
		{
			name:     "inherit mode",
			defaults: model.Defaults{WatchMode: registry.WatchModeSemver},
			want:     registry.WatchModeSemver,
		},
		{
			name:     "legacy item overrides default mode",
			fields:   "  watch_repo: false\n",
			defaults: model.Defaults{WatchMode: registry.WatchModeAll},
			want:     registry.WatchModeTag,
		},
		{
			name:     "mode overrides legacy default",
			fields:   "  watch_mode: tag\n",
			defaults: model.Defaults{WatchRepo: new(true)},
			want:     registry.WatchModeTag,
		},
		{
			name:   "mode wins over legacy item",
			fields: "  watch_repo: true\n  watch_mode: semver\n",
			want:   registry.WatchModeSemver,
		},
		{
			name:    "invalid mode",
			fields:  "  watch_mode: invalid\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "images.yml")
			require.NoError(t, os.WriteFile(filename, []byte("- name: alpine:3.20.0\n"+tt.fields), 0o600))
			jobs := New(&model.PrdFile{Filename: filename}, &tt.defaults).ListJob()
			if tt.wantErr {
				assert.Empty(t, jobs)
				return
			}
			require.Len(t, jobs, 1)
			assert.Equal(t, tt.want, jobs[0].Image.WatchMode)
		})
	}
}

func TestListJobRejectsUnknownFields(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "images.yml")
	require.NoError(t, os.WriteFile(filename, []byte("- name: alpine\n  unknown: value\n"), 0o600))

	fc := New(&model.PrdFile{
		Filename: filename,
	}, &defaults)

	assert.Empty(t, fc.ListJob())
}

func TestListJobAssumeImmutableTags(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "images.yml")
	require.NoError(t, os.WriteFile(filename, []byte("- name: alpine\n- name: busybox\n  assume_immutable_tags: false\n- name: ubuntu\n  assume_immutable_tags: true\n"), 0o600))
	for _, inherited := range []*bool{nil, new(true), new(false)} {
		imageDefaults := model.Defaults{
			SortTags:            registry.SortTagReverse,
			AssumeImmutableTags: inherited,
		}
		jobs := New(&model.PrdFile{Filename: filename}, &imageDefaults).ListJob()
		require.Len(t, jobs, 3)
		require.Equal(t, inherited, jobs[0].Image.AssumeImmutableTags)
		require.Equal(t, new(false), jobs[1].Image.AssumeImmutableTags)
		require.Equal(t, new(true), jobs[2].Image.AssumeImmutableTags)
	}
}

func TestListJobTagFilters(t *testing.T) {
	tests := []struct {
		name        string
		defaultTags []string
		imageTags   []string
		expected    []string
	}{
		{
			name: "no excludes",
		},
		{
			name:        "empty excludes",
			defaultTags: []string{},
			imageTags:   []string{},
		},
		{
			name:        "default excludes only",
			defaultTags: []string{"SNAPSHOT"},
			expected:    []string{"SNAPSHOT"},
		},
		{
			name:      "image excludes only",
			imageTags: []string{"^8\\.0\\..*"},
			expected:  []string{"^8\\.0\\..*"},
		},
		{
			name:        "defaults precede image excludes",
			defaultTags: []string{"SNAPSHOT", "beta"},
			imageTags:   []string{"^8\\.0\\..*", "latest"},
			expected:    []string{"SNAPSHOT", "beta", "^8\\.0\\..*", "latest"},
		},
		{
			name:        "duplicate excludes keep first occurrence",
			defaultTags: []string{"SNAPSHOT", "beta", "SNAPSHOT"},
			imageTags:   []string{"beta", "latest", "latest", "SNAPSHOT"},
			expected:    []string{"SNAPSHOT", "beta", "latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "images.yml")
			data, err := yaml.Marshal([]model.Image{
				{
					Name:        "alpine",
					IncludeTags: []string{`^1\.2\..*`},
					ExcludeTags: tt.imageTags,
				},
				{
					Name: "traefik",
				},
			})
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filename, data, 0o600))

			// Spare capacity exposes accidental appends into the defaults' backing array.
			var defaultTags []string
			if tt.defaultTags != nil {
				defaultTags = make([]string, len(tt.defaultTags), len(tt.defaultTags)+len(tt.imageTags)+1)
				copy(defaultTags, tt.defaultTags)
			}
			imageDefaults := model.Defaults{
				SortTags:    registry.SortTagReverse,
				IncludeTags: []string{`^\d+\.\d+\.\d+$`},
				ExcludeTags: defaultTags,
			}
			fc := New(&model.PrdFile{Filename: filename}, &imageDefaults)

			jobs := fc.ListJob()
			require.Len(t, jobs, 2)
			assert.Equal(t, tt.expected, jobs[0].Image.ExcludeTags)
			assert.Equal(t, []string{`^1\.2\..*`}, jobs[0].Image.IncludeTags)
			assert.Equal(t, imageDefaults.IncludeTags, jobs[1].Image.IncludeTags)
			assert.Equal(t, tt.defaultTags, imageDefaults.ExcludeTags)

			// Changing one job must not affect another job, defaults, or a later read.
			if len(jobs[0].Image.ExcludeTags) > 0 {
				jobs[0].Image.ExcludeTags[0] = "changed"
			}
			assert.Equal(t, tt.defaultTags, imageDefaults.ExcludeTags)
			nextJobs := fc.ListJob()
			require.Len(t, nextJobs, 2)
			assert.Equal(t, tt.expected, nextJobs[0].Image.ExcludeTags)
			assert.Equal(t, nextJobs[1].Image.ExcludeTags, jobs[1].Image.ExcludeTags)
		})
	}
}
