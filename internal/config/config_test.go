package config_test

import (
	"testing"

	"github.com/crazy-max/diun/internal/config"
	"github.com/crazy-max/diun/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		name     string
		flags    model.Flags
		wantData *config.Config
		wantErr  bool
	}{
		{
			name:    "Fail on non-existing file",
			flags:   model.Flags{},
			wantErr: true,
		},
		{
			name: "Fail on wrong file format",
			flags: model.Flags{
				Cfgfile: "config.invalid.yml",
			},
			wantErr: true,
		},
		{
			name: "Success",
			flags: model.Flags{
				Cfgfile: "config.test.yml",
			},
			wantData: &config.Config{
				Flags: model.Flags{
					Cfgfile: "config.test.yml",
				},
				App: model.App{
					ID:      "diun",
					Name:    "Diun",
					Desc:    "Docker image update notifier",
					URL:     "https://github.com/crazy-max/diun",
					Author:  "CrazyMax",
					Version: "test",
				},
				Db: model.Db{
					Path: "diun.db",
				},
				Watch: model.Watch{
					Workers:  100,
					Schedule: "*/30 * * * *",
				},
				Notif: model.Notif{
					Mail: model.Mail{
						Enable:             false,
						Host:               "localhost",
						Port:               25,
						SSL:                false,
						InsecureSkipVerify: false,
					},
					Webhook: model.Webhook{
						Enable:   false,
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"Content-Type":  "application/json",
							"Authorization": "Token123456",
						},
						Timeout: 10,
					},
				},
				RegOpts: map[string]model.RegOpts{
					"someregopts": {
						Timeout: 5,
					},
					"bintrayoptions": {
						Username: "foo",
						Password: "bar",
					},
				},
				Providers: model.Providers{
					Docker: []model.PrdDocker{
						{
							ID:             "local",
							WatchByDefault: true,
						},
					},
					Image: []model.PrdImage{
						{
							Name:      "docker.io/crazymax/nextcloud:latest",
							Os:        "linux",
							Arch:      "amd64",
							RegOptsID: "someregopts",
						},
						{
							Name:      "crazymax/swarm-cronjob",
							Os:        "linux",
							Arch:      "amd64",
							WatchRepo: true,
							IncludeTags: []string{
								`^1\.2\..*`,
							},
						},
						{
							Name:      "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
							Os:        "linux",
							Arch:      "amd64",
							RegOptsID: "bintrayoptions",
						},
						{
							Name:      "docker.bintray.io/jfrog/xray-server:2.8.6",
							Os:        "linux",
							Arch:      "amd64",
							WatchRepo: true,
							MaxTags:   50,
						},
						{
							Name: "quay.io/coreos/hyperkube",
							Os:   "linux",
							Arch: "amd64",
						},
						{
							Name:      "docker.io/portainer/portainer",
							Os:        "linux",
							Arch:      "amd64",
							WatchRepo: true,
							MaxTags:   10,
							IncludeTags: []string{
								`^(0|[1-9]\d*)\..*`,
							},
						},
						{
							Name:      "traefik",
							Os:        "linux",
							Arch:      "amd64",
							WatchRepo: true,
						},
						{
							Name: "alpine",
							Os:   "linux",
							Arch: "arm64v8",
						},
						{
							Name: "docker.io/graylog/graylog:3.2.0",
							Os:   "linux",
							Arch: "amd64",
						},
						{
							Name: "jacobalberty/unifi:5.9",
							Os:   "linux",
							Arch: "amd64",
						},
						{
							Name: "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
							Os:   "linux",
							Arch: "amd64",
						},
						{
							Name:      "crazymax/ddns-route53",
							Os:        "linux",
							Arch:      "amd64",
							WatchRepo: true,
							IncludeTags: []string{
								`^1\..*`,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.flags, "test")
			assert.Equal(t, tt.wantData, cfg)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
