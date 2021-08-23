package file_test

import (
	"testing"

	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/internal/provider/file"
	"github.com/stretchr/testify/assert"
)

var (
	bintrayFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:     "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
				RegOpt:   "bintrayoptions",
				NotifyOn: model.NotifyOnDefaults,
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.bintray.io/jfrog/xray-server:2.8.6",
				WatchRepo: true,
				NotifyOn: []model.NotifyOn{
					model.NotifyOnNew,
				},
				MaxTags: 50,
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
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/swarm-cronjob",
				WatchRepo: true,
				NotifyOn:  model.NotifyOnDefaults,
				IncludeTags: []string{
					`^1\.2\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.io/portainer/portainer",
				WatchRepo: true,
				NotifyOn:  model.NotifyOnDefaults,
				MaxTags:   10,
				IncludeTags: []string{
					`^(0|[1-9]\d*)\..*`,
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "traefik",
				WatchRepo: true,
				NotifyOn:  model.NotifyOnDefaults,
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:     "alpine",
				NotifyOn: model.NotifyOnDefaults,
				Platform: model.ImagePlatform{
					OS:      "linux",
					Arch:    "arm64",
					Variant: "v8",
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:     "docker.io/graylog/graylog:3.2.0",
				NotifyOn: model.NotifyOnDefaults,
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:     "jacobalberty/unifi:5.9",
				NotifyOn: model.NotifyOnDefaults,
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/ddns-route53",
				WatchRepo: true,
				NotifyOn:  model.NotifyOnDefaults,
				IncludeTags: []string{
					`^1\..*`,
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
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:     "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
				NotifyOn: model.NotifyOnDefaults,
			},
		},
	}
)

func TestListJobFilename(t *testing.T) {
	fc := file.New(&model.PrdFile{
		Filename: "./fixtures/dockerhub.yml",
	})
	assert.Equal(t, dockerhubFile, fc.ListJob())
}

func TestListJobDirectory(t *testing.T) {
	fc := file.New(&model.PrdFile{
		Directory: "./fixtures",
	})
	assert.Equal(t, append(append(bintrayFile, dockerhubFile...), quayFile...), fc.ListJob())
}
