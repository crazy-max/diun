package config_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v3/internal/config"
	"github.com/crazy-max/diun/v3/internal/model"
	"github.com/crazy-max/diun/v3/pkg/traefik/config/env"
	"github.com/crazy-max/diun/v3/pkg/utl"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestLoad(t *testing.T) {
	defaultCfg := config.Config(model.DefaultConfig)
	cases := []struct {
		name     string
		cli      model.Cli
		wantData *config.Config
		wantErr  bool
	}{
		{
			name:     "Default on non-existing file",
			cli:      model.Cli{},
			wantData: &defaultCfg,
			wantErr:  false,
		},
		{
			name: "Fail on wrong file format",
			cli: model.Cli{
				Cfgfile: "./fixtures/config.invalid.yml",
			},
			wantErr: true,
		},
		{
			name: "Success",
			cli: model.Cli{
				Cfgfile: "./fixtures/config.test.yml",
			},
			wantData: &config.Config{
				Db: model.Db{
					Path: "diun.db",
				},
				Watch: model.Watch{
					Workers:         100,
					Schedule:        "*/30 * * * *",
					FirstCheckNotif: utl.NewFalse(),
				},
				Notif: &model.Notif{
					Amqp: &model.NotifAmqp{
						Host:     "localhost",
						Port:     5672,
						Username: "guest",
						Password: "guest",
						Queue:    "queue",
					},
					Gotify: &model.NotifGotify{
						Endpoint: "http://gotify.foo.com",
						Token:    "Token123456",
						Priority: 1,
						Timeout:  utl.NewDuration(10 * time.Second),
					},
					Mail: &model.NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						From:               "diun@example.com",
						To:                 "webmaster@example.com",
					},
					RocketChat: &model.NotifRocketChat{
						Endpoint: "http://rocket.foo.com:3000",
						Channel:  "#general",
						UserID:   "abcdEFGH012345678",
						Token:    "Token123456",
						Timeout:  utl.NewDuration(10 * time.Second),
					},
					Script: &model.NotifScript{
						Cmd: "go",
						Args: []string{
							"version",
						},
					},
					Slack: &model.NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Teams: &model.NotifTeams{
						WebhookURL: "https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Telegram: &model.NotifTelegram{
						Token:   "abcdef123456",
						ChatIDs: []int64{8547439, 1234567},
					},
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "application/json",
							"authorization": "Token123456",
						},
						Timeout: utl.NewDuration(10 * time.Second),
					},
				},
				RegOpts: map[string]model.RegOpts{
					"someregopts": {
						InsecureTLS: utl.NewFalse(),
						Timeout:     utl.NewDuration(5 * time.Second),
					},
					"bintrayoptions": {
						Username:    "foo",
						Password:    "bar",
						InsecureTLS: utl.NewFalse(),
						Timeout:     utl.NewDuration(10 * time.Second),
					},
					"sensitive": {
						UsernameFile: "/run/secrets/username",
						PasswordFile: "/run/secrets/password",
						InsecureTLS:  utl.NewFalse(),
						Timeout:      utl.NewDuration(10 * time.Second),
					},
				},
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewTrue(),
						WatchStopped:   utl.NewTrue(),
					},
					Swarm: &model.PrdSwarm{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewTrue(),
					},
					File: &model.PrdFile{
						Filename: "./fixtures/dummy.yml",
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
			ex, _ := yaml.Marshal(tt.wantData)
			ac, _ := yaml.Marshal(cfg)
			assert.Equal(t, string(ex), string(ac))
			if !tt.wantErr && cfg != nil {
				assert.NotEmpty(t, cfg.Display())
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	defer UnsetEnv("DIUN_")()

	testConfig := config.Config{
		Db: model.Db{
			Path: "diunenv.db",
		},
		Watch: model.Watch{
			Workers:         32,
			Schedule:        "* * * * *",
			FirstCheckNotif: utl.NewTrue(),
		},
		Notif: &model.Notif{
			Amqp: &model.NotifAmqp{
				Host:     "127.0.0.1",
				Port:     56720,
				Username: "guestwhat",
				Password: "guestwhat",
				Queue:    "queue2",
			},
			Gotify: &model.NotifGotify{
				Endpoint: "http://gotify.example.com",
				Token:    "Token123456789",
				Priority: 2,
				Timeout:  utl.NewDuration(20 * time.Second),
			},
			Mail: &model.NotifMail{
				Host:               "127.0.0.1",
				Port:               25,
				SSL:                utl.NewTrue(),
				InsecureSkipVerify: utl.NewTrue(),
				From:               "diun@foo.com",
				To:                 "webmaster@foo.com",
			},
			RocketChat: &model.NotifRocketChat{
				Endpoint: "http://rocket.example.com",
				Channel:  "#diun",
				UserID:   "abcd1234",
				Token:    "Token123456789",
				Timeout:  utl.NewDuration(10 * time.Second),
			},
			Script: &model.NotifScript{
				Cmd: "go",
				Args: []string{
					"version",
				},
			},
			Slack: &model.NotifSlack{
				WebhookURL: "https://hooks.slack.com/services/AB1234/HIJK34LMN/01234567890abcdefghij",
			},
			Teams: &model.NotifTeams{
				WebhookURL: "https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
			},
			Telegram: &model.NotifTelegram{
				Token:   "abcdef123456",
				ChatIDs: []int64{1234567, 891012},
			},
			Webhook: &model.NotifWebhook{
				Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
				Method:   "GET",
				Headers: map[string]string{
					"content-type":  "text/plain",
					"authorization": "Token78910",
				},
				Timeout: utl.NewDuration(20 * time.Second),
			},
		},
		RegOpts: map[string]model.RegOpts{
			"someregopts": {
				Timeout: utl.NewDuration(20 * time.Second),
			},
			"bintrayoptions": {
				Username: "foo",
				Password: "bar5",
			},
			"sensitive": {
				UsernameFile: "/run/secrets/username2",
				PasswordFile: "/run/secrets/password3",
			},
		},
		Providers: &model.Providers{
			Docker: &model.PrdDocker{
				TLSVerify:      utl.NewFalse(),
				WatchByDefault: utl.NewFalse(),
				WatchStopped:   utl.NewFalse(),
			},
			Swarm: &model.PrdSwarm{
				TLSVerify:      utl.NewFalse(),
				WatchByDefault: utl.NewFalse(),
			},
			File: &model.PrdFile{
				Filename: "./fixtures/dummy.yml",
			},
		},
	}

	dec, err := env.Encode(&testConfig)
	for _, value := range dec {
		os.Setenv(strings.Replace(value.Name, "TRAEFIK_", "DIUN_", 1), value.Default)
		//fmt.Println(fmt.Sprintf(`%s=%s`, strings.Replace(value.Name, "TRAEFIK_", "DIUN_", 1), value.Default))
	}

	cfg, err := config.Load(model.Cli{
		Cfgfile: "./fixtures/config.test.yml",
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	ex, _ := yaml.Marshal(testConfig)
	ac, _ := yaml.Marshal(cfg)
	assert.Equal(t, string(ex), string(ac))
}

func UnsetEnv(prefix string) (restore func()) {
	before := map[string]string{}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]

		os.Unsetenv(parts[0])
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}

			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			// Check if the envar previously existed
			v, ok := before[parts[0]]
			if !ok {
				// This is a newly added envar with prefix, zap it
				os.Unsetenv(parts[0])
				continue
			}

			if parts[1] != v {
				// If the envar value has changed, set it back
				os.Setenv(parts[0], v)
			}
		}

		// Still need to check if there have been any deleted envars
		for k, v := range before {
			if _, ok := after[k]; !ok {
				// k is not present in after, so we set it.
				os.Setenv(k, v)
			}
		}
	}
}
