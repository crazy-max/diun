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
		cli      model.Cli
		wantData *config.Config
		wantErr  bool
	}{
		{
			name:    "Fail on non-existing file",
			cli:     model.Cli{},
			wantErr: true,
		},
		{
			name: "Fail on wrong file format",
			cli: model.Cli{
				Cfgfile: "config.invalid.yml",
			},
			wantErr: true,
		},
		{
			name: "Success",
			cli: model.Cli{
				Cfgfile: "config.test.yml",
			},
			wantData: &config.Config{
				Cli: model.Cli{
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
					RocketChat: model.NotifRocketChat{
						Enable:   false,
						Endpoint: "http://rocket.foo.com:3000",
						Channel:  "#general",
						UserID:   "abcdEFGH012345678",
						Token:    "Token123456",
						Timeout:  10,
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
					File: model.PrdFile{
						Filename: "./dummy.yml",
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.cli, "test")
			if !tt.wantErr && err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.wantData, cfg)
		})
	}
}
