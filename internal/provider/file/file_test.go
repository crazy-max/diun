package file

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/stretchr/testify/assert"
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
				Name:     "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
				RegOpt:   "bintrayoptions",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				WatchRepo: model.WatchRepoAll,
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
				Name:     "docker.io/crazymax/nextcloud:latest",
				RegOpt:   "myregistry",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				WatchRepo: model.WatchRepoAll,
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
				WatchRepo: model.WatchRepoAll,
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
				WatchRepo: model.WatchRepoAll,
				NotifyOn:  model.NotifyOnDefaults,
				SortTags:  registry.SortTagDefault,
				MaxTags:   25,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
				ExcludeTags: []string{
					`latest`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:     "alpine",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
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
				Name:     "docker.io/graylog/graylog:3.2.0",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				Name:     "jacobalberty/unifi:5.9",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				WatchRepo: model.WatchRepoAll,
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
				Name:     "quay.io/coreos/hyperkube",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				Name:     "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				MaxTags:  25,
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
				Name:     "lscr.io/linuxserver/heimdall",
				NotifyOn: model.NotifyOnDefaults,
				SortTags: registry.SortTagReverse,
				HubLink:  "https://fleet.linuxserver.io/image?name=linuxserver/heimdall",
				MaxTags:  25,
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
		WatchRepo: model.WatchRepoAll,
	})

	for _, job := range fc.ListJob() {
		assert.Equal(t, model.WatchRepoAll, job.Image.WatchRepo)
	}
}
