package file_test

import (
	"testing"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/internal/provider/file"
	"github.com/stretchr/testify/assert"
)

var (
	bintrayFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
				RegOptsID: "bintrayoptions",
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.bintray.io/jfrog/xray-server:2.8.6",
				WatchRepo: true,
				MaxTags:   50,
			},
		},
	}
	dockerhubFile = []model.Job{
		{
			Provider: "file",
			Image: model.Image{
				Name:      "docker.io/crazymax/nextcloud:latest",
				RegOptsID: "someregopts",
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/swarm-cronjob",
				WatchRepo: true,
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
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name: "alpine",
				Platform: model.ImagePlatform{
					Os:      "linux",
					Arch:    "arm64",
					Variant: "v8",
				},
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name: "docker.io/graylog/graylog:3.2.0",
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name: "jacobalberty/unifi:5.9",
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name:      "crazymax/ddns-route53",
				WatchRepo: true,
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
				Name: "quay.io/coreos/hyperkube",
			},
		},
		{
			Provider: "file",
			Image: model.Image{
				Name: "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
			},
		},
	}
)

func TestListJobFilename(t *testing.T) {
	fc := file.New(&model.PrdFile{
		Filename: "./test/dockerhub.yml",
	})
	assert.Equal(t, dockerhubFile, fc.ListJob())
}

func TestListJobDirectory(t *testing.T) {
	fc := file.New(&model.PrdFile{
		Directory: "./test",
	})
	assert.Equal(t, append(append(bintrayFile, dockerhubFile...), quayFile...), fc.ListJob())
}
