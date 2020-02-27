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
					Gotify: model.NotifGotify{
						Enable:   false,
						Endpoint: "http://gotify.foo.com",
						Token:    "Token123456",
						Priority: 1,
						Timeout:  10,
					},
					Mail: model.NotifMail{
						Enable:             false,
						Host:               "localhost",
						Port:               25,
						SSL:                false,
						InsecureSkipVerify: false,
					},
					Slack: model.NotifSlack{
						Enable:     false,
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Telegram: model.NotifTelegram{
						Enable:   false,
						BotToken: "abcdef123456",
						ChatIDs:  []int64{8547439, 1234567},
					},
					Webhook: model.NotifWebhook{
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
					"sensitive": {
						UsernameFile: "/run/secrets/username",
						PasswordFile: "/run/secrets/password",
					},
				},
				Providers: model.Providers{
					Docker: map[string]model.PrdDocker{
						"standalone": {
							TLSVerify:      true,
							WatchByDefault: true,
							WatchStopped:   true,
						},
					},
					Swarm: map[string]model.PrdSwarm{
						"local_swarm": {
							TLSVerify:      true,
							WatchByDefault: true,
						},
					},
					Static: []model.PrdStatic{
						{
							Name:      "docker.io/crazymax/nextcloud:latest",
							RegOptsID: "someregopts",
						},
						{
							Name:      "crazymax/swarm-cronjob",
							WatchRepo: true,
							IncludeTags: []string{
								`^1\.2\..*`,
							},
						},
						{
							Name:      "jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0",
							RegOptsID: "bintrayoptions",
						},
						{
							Name:      "docker.bintray.io/jfrog/xray-server:2.8.6",
							WatchRepo: true,
							MaxTags:   50,
						},
						{
							Name: "quay.io/coreos/hyperkube",
						},
						{
							Name:      "docker.io/portainer/portainer",
							WatchRepo: true,
							MaxTags:   10,
							IncludeTags: []string{
								`^(0|[1-9]\d*)\..*`,
							},
						},
						{
							Name:      "traefik",
							WatchRepo: true,
						},
						{
							Name: "alpine",
							Os:   "linux",
							Arch: "arm64v8",
						},
						{
							Name: "docker.io/graylog/graylog:3.2.0",
						},
						{
							Name: "jacobalberty/unifi:5.9",
						},
						{
							Name: "quay.io/coreos/hyperkube:v1.1.7-coreos.1",
						},
						{
							Name:      "crazymax/ddns-route53",
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
