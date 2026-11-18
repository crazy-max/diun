package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/gonfig/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAssumeImmutableTags(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "diun.yml")
	require.NoError(t, os.WriteFile(filename, []byte("providers:\n  docker: {}\ndefaults:\n  assumeImmutableTags: true\n"), 0o600))
	cfg, err := Load(filename)
	require.NoError(t, err)
	require.Equal(t, new(true), cfg.Defaults.AssumeImmutableTags)
	t.Setenv("DIUN_DEFAULTS_ASSUMEIMMUTABLETAGS", "false")
	cfg, err = Load(filename)
	require.NoError(t, err)
	require.Equal(t, new(false), cfg.Defaults.AssumeImmutableTags)
}

func TestLoadNtfyHeadersFromEnv(t *testing.T) {
	t.Setenv("DIUN_NOTIF_NTFY_TOPIC", "diun")
	t.Setenv("DIUN_NOTIF_NTFY_HEADERS_P-ACCESS-TOKEN-ID", "proxy-id")
	t.Setenv("DIUN_NOTIF_NTFY_HEADERS_P-ACCESS-TOKEN", "proxy-token")

	cfg, err := Load("./fixtures/config.file.yml")
	require.NoError(t, err)
	require.NotNil(t, cfg.Notif.Ntfy)
	require.Equal(t, map[string]string{
		"p-access-token-id": "proxy-id",
		"p-access-token":    "proxy-token",
	}, cfg.Notif.Ntfy.Headers)
}

func TestLoadMatrixToken(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("file-token"), 0o600))
	for _, source := range []string{"file", "env"} {
		t.Run(source, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "diun.yml")
			contents := "providers:\n  docker: {}\nnotif:\n  matrix:\n    roomID: '!alerts:example.com'\n"
			if source == "file" {
				contents += "    token: inline-token\n    tokenFile: '" + tokenFile + "'\n"
			} else {
				t.Setenv("DIUN_NOTIF_MATRIX_TOKEN", "inline-token")
				t.Setenv("DIUN_NOTIF_MATRIX_TOKENFILE", tokenFile)
			}
			require.NoError(t, os.WriteFile(filename, []byte(contents), 0o600))
			t.Setenv("DIUN_DB_PATH", filepath.Join(t.TempDir(), "diun.db"))
			cfg, err := Load(filename)
			require.NoError(t, err)
			require.NotNil(t, cfg.Notif.Matrix)
			assert.Equal(t, "inline-token", cfg.Notif.Matrix.Token)
			assert.Equal(t, tokenFile, cfg.Notif.Matrix.TokenFile)
			assert.Empty(t, cfg.Notif.Matrix.User)
			assert.Empty(t, cfg.Notif.Matrix.Password)
		})
	}
}

func TestLoadFile(t *testing.T) {
	cases := []struct {
		name     string
		cfg      string
		wantData *Config
		wantErr  bool
	}{
		{
			name:    "Failed on non-existing file",
			cfg:     "",
			wantErr: true,
		},
		{
			name:    "Fail on wrong file format",
			cfg:     "./fixtures/config.invalid.yml",
			wantErr: true,
		},
		{
			name:    "Fail on no UUID for Healthchecks",
			cfg:     "./fixtures/config.err.hc.yml",
			wantErr: true,
		},
		{
			name: "Success with healthchecks uuidFile",
			cfg:  "./fixtures/config.hc.uuidfile.yml",
			wantData: &Config{
				Db: (&model.Db{}).GetDefaults(),
				Watch: &model.Watch{
					Workers:         10,
					Jitter:          new(30 * time.Second),
					FirstCheckNotif: new(false),
					RunOnStartup:    new(true),
					CompareDigest:   new(true),
					Healthchecks: &model.Healthchecks{
						BaseURL:  "https://hc-ping.com/",
						UUIDFile: "./fixtures/run_secrets_uuid",
					},
				},
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
		},
		{
			name:    "Fail on no provider",
			cfg:     "./fixtures/config.err.provider.yml",
			wantErr: true,
		},
		{
			name:    "Fail on invalid metrics metadata label",
			cfg:     "./fixtures/config.err.metrics.yml",
			wantErr: true,
		},
		{
			name: "Success",
			cfg:  "./fixtures/config.test.yml",
			wantData: &Config{
				Db: &model.Db{
					Path: "diun.db",
				},
				Watch: &model.Watch{
					Workers:         100,
					Schedule:        "*/30 * * * *",
					Jitter:          new(30 * time.Second),
					FirstCheckNotif: new(true),
					RunOnStartup:    new(false),
					CompareDigest:   new(true),
					Healthchecks: &model.Healthchecks{
						BaseURL: "https://hc-ping.com/",
						UUID:    "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
					},
				},
				Defaults: &model.Defaults{
					NotifyOn: []model.NotifyOn{model.NotifyOnNew},
					MaxTags:  5,
					SortTags: registry.SortTagReverse,
				},
				Metrics: (&model.Metrics{}).GetDefaults(),
				Notif: &model.Notif{
					Amqp: &model.NotifAmqp{
						Host:     "localhost",
						Port:     5672,
						Username: "guest",
						Password: "guest",
						Queue:    "queue",
					},
					Apprise: &model.NotifApprise{
						Endpoint:      "http://apprise:8000",
						Token:         "abc",
						Tags:          []string{"diun"},
						Timeout:       new(10 * time.Second),
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody:  model.NotifDefaultTemplateBody,
					},
					Discord: &model.NotifDiscord{
						WebhookURL: "https://discordapp.com/api/webhooks/1234567890/Abcd-eFgh-iJklmNo_pqr",
						Mentions: []string{
							"@here",
							"@everyone",
							"<@124>",
							"<@125>",
							"<@&200>",
						},
						RenderEmbeds: new(true),
						RenderFields: new(true),
						Timeout:      new(10 * time.Second),
						TemplateBody: model.NotifDefaultTemplateBody,
					},
					Elasticsearch: &model.NotifElasticsearch{
						Address:  "https://elastic.foo.com",
						Username: "elastic",
						Password: "password",
						Client:   "diun",
						Index:    "diun-notifications",
						Timeout:  new(10 * time.Second),
					},
					Gotify: &model.NotifGotify{
						Endpoint:      "http://gotify.foo.com",
						Token:         "Token123456",
						Priority:      1,
						Timeout:       new(10 * time.Second),
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody:  model.NotifDefaultTemplateBody,
					},
					Mail: &model.NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                new(false),
						InsecureSkipVerify: new(false),
						From:               "diun@example.com",
						To: []string{
							"webmaster@example.com",
							"me@example.com",
						},
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody: `Docker tag {{ if .Entry.Image.HubLink }}[**{{ .Entry.Image }}**]({{ .Entry.Image.HubLink }}){{ else }}**{{ .Entry.Image }}**{{ end }}
which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}.

This image has been {{ if (eq .Entry.Status "new") }}created{{ else }}updated{{ end }} at
<code>{{ .Entry.Manifest.Created.Format "Jan 02, 2006 15:04:05 UTC" }}</code> with digest <code>{{ .Entry.Manifest.Digest }}</code>
for <code>{{ .Entry.Manifest.Platform }}</code> platform.
`,
					},
					Matrix: &model.NotifMatrix{
						HomeserverURL: "https://matrix.org",
						User:          "@foo:matrix.org",
						Password:      "bar",
						RoomID:        "!abcdefGHIjklmno:matrix.org",
						MsgType:       model.NotifMatrixMsgTypeNotice,
						TemplateBody:  model.NotifDefaultTemplateBody,
					},
					Mqtt: &model.NotifMqtt{
						Scheme:   "mqtt",
						Host:     "localhost",
						Port:     1883,
						Username: "guest",
						Password: "guest",
						Client:   "diun",
						Topic:    "docker/diun",
						QoS:      0,
					},
					Ntfy: &model.NotifNtfy{
						Endpoint: "https://ntfy.sh",
						Headers: map[string]string{
							"P-Access-Token-Id": "test-id",
							"P-Access-Token":    "test-token",
						},
						Topic:         "diun-acce65a0-b777-46f9-9a11-58c67d1579c4",
						Priority:      3,
						Tags:          []string{"package"},
						Icon:          "https://example.com/diun.png",
						Click:         "https://dock.example.com/compose/{{ .Entry.Metadata.stack }}",
						Timeout:       new(10 * time.Second),
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody:  model.NotifDefaultTemplateBody,
					},
					Pushover: &model.NotifPushover{ //nolint:gosec // fixture values are test data.
						Token:         "uQiRzpo4DXghDmr9QzzfQu27cmVRsG",
						Recipient:     "gznej3rKEVAvPUxu9vvNnqpmZpokzF",
						Timeout:       new(10 * time.Second),
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody:  model.NotifDefaultTemplateBody,
					},
					RocketChat: &model.NotifRocketChat{
						Endpoint:         "http://rocket.foo.com:3000",
						Channel:          "#general",
						UserID:           "abcdEFGH012345678",
						Token:            "Token123456",
						RenderAttachment: new(true),
						Timeout:          new(10 * time.Second),
						TemplateTitle:    model.NotifDefaultTemplateTitle,
						TemplateBody:     model.NotifRocketChatDefaultTemplateBody,
					},
					Script: &model.NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
					Slack: &model.NotifSlack{
						WebhookURL:   "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
						RenderFields: new(false),
						TemplateBody: model.NotifSlackDefaultTemplateBody,
					},
					Teams: &model.NotifTeams{
						WebhookURL:   "https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
						CardType:     model.NotifTeamsCardTypeMessageCard,
						RenderFacts:  new(false),
						Timeout:      new(10 * time.Second),
						TemplateBody: model.NotifTeamsDefaultTemplateBody,
					},
					Telegram: &model.NotifTelegram{
						APIURL: gotgbot.DefaultAPIURL,
						Token:  "abcdef123456",
						Proxy:  "http://proxy.foo.com:3128",
						ChatIDs: []string{
							"8547439",
							"1234567",
							"567891234:25",
							"891256734:25;12",
						},
						TemplateBody:        model.NotifTelegramDefaultTemplateBody,
						DisableNotification: new(false),
					},
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "application/json",
							"authorization": "Token123456",
						},
						Timeout: new(10 * time.Second),
					},
				},
				RegOpts: model.RegOpts{
					model.RegOpt{
						Name:        "myregistry",
						Selector:    model.RegOptSelectorName,
						Username:    "fii",
						Password:    "bor",
						InsecureTLS: new(false),
						Timeout:     new(5 * time.Second),
					},
					model.RegOpt{
						Name:        "docker.io",
						Selector:    model.RegOptSelectorImage,
						Username:    "foo",
						Password:    "bar",
						InsecureTLS: new(false),
						Timeout:     new(time.Duration(0)),
					},
					model.RegOpt{ //nolint:gosec // fixture paths are test data.
						Name:         "docker.io/crazymax",
						Selector:     model.RegOptSelectorImage,
						UsernameFile: "./fixtures/run_secrets_username",
						PasswordFile: "./fixtures/run_secrets_password",
						InsecureTLS:  new(false),
						Timeout:      new(time.Duration(0)),
					},
				},
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(true),
						WatchStopped:   new(true),
					},
					Swarm: &model.PrdSwarm{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
					},
					Kubernetes: &model.PrdKubernetes{
						TLSInsecure:    new(false),
						WatchByDefault: new(true),
					},
					File: &model.PrdFile{
						Filename: "./fixtures/file.yml",
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, cfg)
		})
	}
}

func TestLoadEnv(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      string
		environ  []string
		expected any
		wantErr  bool
	}{
		{
			desc:     "no env vars",
			environ:  nil,
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "healthchecks with UUIDFILE env var",
			environ: []string{
				"DIUN_WATCH_HEALTHCHECKS_UUIDFILE=./fixtures/run_secrets_uuid",
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &Config{
				Db: (&model.Db{}).GetDefaults(),
				Watch: &model.Watch{
					Workers:         10,
					Jitter:          new(30 * time.Second),
					FirstCheckNotif: new(false),
					RunOnStartup:    new(true),
					CompareDigest:   new(true),
					Healthchecks: &model.Healthchecks{
						UUIDFile: "./fixtures/run_secrets_uuid",
					},
				},
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "docker provider",
			environ: []string{
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif:    nil,
				RegOpts:  nil,
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "containerd provider",
			environ: []string{
				"DIUN_PROVIDERS_CONTAINERD=true",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif:    nil,
				RegOpts:  nil,
				Providers: &model.Providers{
					Containerd: &model.PrdContainerd{
						Namespaces:     []string{"default"},
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "nomad provider namespaces",
			environ: []string{
				"DIUN_PROVIDERS_NOMAD=true",
				"DIUN_PROVIDERS_NOMAD_NAMESPACE=legacy",
				"DIUN_PROVIDERS_NOMAD_NAMESPACES=dev,prod",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif:    nil,
				RegOpts:  nil,
				Providers: &model.Providers{
					Nomad: &model.PrdNomad{
						Namespace:      "legacy", //nolint:staticcheck // Test backward compatibility with the deprecated field.
						Namespaces:     []string{"dev", "prod"},
						TLSInsecure:    new(false),
						WatchByDefault: new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "metrics and docker provider",
			environ: []string{
				"DIUN_METRICS_ENABLED=true",
				"DIUN_METRICS_ADDR=127.0.0.1:9090",
				"DIUN_METRICS_PATH=/prometheus",
				"DIUN_METRICS_TOKENFILE=./fixtures/run_secrets_username",
				"DIUN_METRICS_METADATALABELS=channel,environment",
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics: &model.Metrics{ //nolint:gosec // fixture path is test data.
					Enabled:   new(true),
					Addr:      "127.0.0.1:9090",
					Path:      "/prometheus",
					TokenFile: "./fixtures/run_secrets_username",
					MetadataLabels: []string{
						"channel",
						"environment",
					},
				},
				Notif:   nil,
				RegOpts: nil,
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "docker provider and regopts",
			environ: []string{
				"DIUN_REGOPTS_0_NAME=docker.io",
				"DIUN_REGOPTS_0_SELECTOR=image",
				"DIUN_REGOPTS_0_USERNAMEFILE=./fixtures/run_secrets_username",
				"DIUN_REGOPTS_0_PASSWORDFILE=./fixtures/run_secrets_password",
				"DIUN_REGOPTS_0_TIMEOUT=30s",
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				RegOpts: model.RegOpts{
					model.RegOpt{ //nolint:gosec // fixture paths are test data.
						Name:         "docker.io",
						Selector:     model.RegOptSelectorImage,
						UsernameFile: "./fixtures/run_secrets_username",
						PasswordFile: "./fixtures/run_secrets_password",
						InsecureTLS:  new(false),
						Timeout:      new(30 * time.Second),
					},
				},
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "swarm provider and notif telegram",
			environ: []string{
				"DIUN_NOTIF_TELEGRAM_TOKEN=abcdef123456",
				"DIUN_NOTIF_TELEGRAM_CHATIDS=8547439,1234567",
				"DIUN_NOTIF_TELEGRAM_APIURL=http://telegram-bot-api:8081",
				"DIUN_NOTIF_TELEGRAM_PROXY=http://proxy.foo.com:3128",
				"DIUN_PROVIDERS_SWARM=true",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif: &model.Notif{
					Telegram: &model.NotifTelegram{
						Token: "abcdef123456",
						ChatIDs: []string{
							"8547439",
							"1234567",
						},
						TemplateBody:        model.NotifTelegramDefaultTemplateBody,
						DisableNotification: new(false),
						APIURL:              "http://telegram-bot-api:8081",
						Proxy:               "http://proxy.foo.com:3128",
					},
				},
				Providers: &model.Providers{
					Swarm: &model.PrdSwarm{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "file provider and notif script",
			environ: []string{
				"DIUN_NOTIF_SCRIPT_CMD=uname",
				"DIUN_NOTIF_SCRIPT_ARGS=-a",
				"DIUN_PROVIDERS_FILE_DIRECTORY=./fixtures",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif: &model.Notif{
					Script: &model.NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
				},
				Providers: &model.Providers{
					File: &model.PrdFile{
						Directory: "./fixtures",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					t.Setenv(n[0], n[1])
				}
			}

			cfg, err := Load(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestLoadMixed(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      string
		environ  []string
		expected any
		wantErr  bool
	}{
		{
			desc: "env vars and invalid file",
			cfg:  "./fixtures/config.invalid.yml",
			environ: []string{
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "docker provider (file) and notif mails (envs)",
			cfg:  "./fixtures/config.docker.yml",
			environ: []string{
				"DIUN_NOTIF_MAIL_HOST=127.0.0.1",
				"DIUN_NOTIF_MAIL_PORT=25",
				"DIUN_NOTIF_MAIL_SSL=false",
				"DIUN_NOTIF_MAIL_INSECURESKIPVERIFY=true",
				"DIUN_NOTIF_MAIL_FROM=diun@foo.com",
				"DIUN_NOTIF_MAIL_TO=webmaster@foo.com",
				"DIUN_NOTIF_MAIL_LOCALNAME=foo.com",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif: &model.Notif{
					Mail: &model.NotifMail{
						Host:               "127.0.0.1",
						Port:               25,
						SSL:                new(false),
						InsecureSkipVerify: new(true),
						LocalName:          "foo.com",
						From:               "diun@foo.com",
						To: []string{
							"webmaster@foo.com",
						},
						TemplateTitle: model.NotifDefaultTemplateTitle,
						TemplateBody: `Docker tag {{ if .Entry.Image.HubLink }}[**{{ .Entry.Image }}**]({{ .Entry.Image.HubLink }}){{ else }}**{{ .Entry.Image }}**{{ end }}
which you subscribed to through {{ .Entry.Provider }} provider has been {{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}.

This image has been {{ if (eq .Entry.Status "new") }}created{{ else }}updated{{ end }} at
<code>{{ .Entry.Manifest.Created.Format "Jan 02, 2006 15:04:05 UTC" }}</code> with digest <code>{{ .Entry.Manifest.Digest }}</code>
for <code>{{ .Entry.Manifest.Platform }}</code> platform.
`,
					},
				},
				RegOpts: nil,
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      new(true),
						WatchByDefault: new(false),
						WatchStopped:   new(false),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "file provider and notif webhook env override",
			cfg:  "./fixtures/config.file.yml",
			environ: []string{
				"DIUN_NOTIF_WEBHOOK_ENDPOINT=http://webhook.foo.com/sd54qad89azd5a",
				"DIUN_NOTIF_WEBHOOK_HEADERS_AUTHORIZATION=Token78910",
				"DIUN_NOTIF_WEBHOOK_HEADERS_CONTENT-TYPE=text/plain",
				"DIUN_NOTIF_WEBHOOK_METHOD=GET",
				"DIUN_NOTIF_WEBHOOK_TIMEOUT=1m",
			},
			expected: &Config{
				Db:       (&model.Db{}).GetDefaults(),
				Watch:    (&model.Watch{}).GetDefaults(),
				Defaults: (&model.Defaults{}).GetDefaults(),
				Metrics:  (&model.Metrics{}).GetDefaults(),
				Notif: &model.Notif{
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "text/plain",
							"authorization": "Token78910",
						},
						Timeout: new(1 * time.Minute),
					},
				},
				RegOpts: nil,
				Providers: &model.Providers{
					File: &model.PrdFile{
						Filename: "./fixtures/file.yml",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					t.Setenv(n[0], n[1])
				}
			}

			cfg, err := Load(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestWatchRepoDeprecationWarning(t *testing.T) {
	tests := []struct {
		name     string
		defaults string
		env      string
		warn     bool
		mode     registry.WatchMode
	}{
		{
			name: "unset",
			mode: registry.WatchModeTag,
		},
		{
			name:     "null defaults",
			defaults: "defaults: null\n",
			mode:     registry.WatchModeTag,
		},
		{
			name:     "modern option",
			defaults: "defaults:\n  watchMode: semver\n",
			mode:     registry.WatchModeSemver,
		},
		{
			name:     "legacy true",
			defaults: "defaults:\n  watchRepo: true\n",
			warn:     true,
			mode:     registry.WatchModeAll,
		},
		{
			name:     "legacy false",
			defaults: "defaults:\n  watchRepo: false\n",
			warn:     true,
			mode:     registry.WatchModeTag,
		},
		{
			name: "environment true",
			env:  "true",
			warn: true,
			mode: registry.WatchModeAll,
		},
		{
			name: "environment false",
			env:  "false",
			warn: true,
			mode: registry.WatchModeTag,
		},
		{
			name:     "both options",
			defaults: "defaults:\n  watchRepo: true\n  watchMode: semver\n",
			warn:     true,
			mode:     registry.WatchModeSemver,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Logger
			log.Logger = zerolog.New(&output)
			t.Cleanup(func() { log.Logger = previous })

			filename := filepath.Join(t.TempDir(), "diun.yml")
			data := tt.defaults + "providers:\n  file:\n    filename: ./fixtures/file.yml\n"
			require.NoError(t, os.WriteFile(filename, []byte(data), 0o600))
			if tt.env != "" {
				t.Setenv("DIUN_DEFAULTS_WATCHREPO", tt.env)
			}
			cfg, err := Load(filename)
			require.NoError(t, err)
			assert.Equal(t, tt.warn, strings.Contains(output.String(), "is deprecated"))
			if tt.warn {
				assert.Contains(t, output.String(), `"level":"warn"`)
				assert.Contains(t, output.String(), "defaults.watchMode / DIUN_DEFAULTS_WATCHMODE")
			}
			image := model.Image{}
			require.NoError(t, image.ResolveWatchMode(cfg.Defaults))
			assert.Equal(t, tt.mode, image.WatchMode)
		})
	}
}

func TestLegacyWatchRepoConfiguration(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "diun.yml")
	data := "defaults:\n  watchRepo: true\nproviders:\n  file:\n    filename: ./fixtures/file.yml\n"
	require.NoError(t, os.WriteFile(filename, []byte(data), 0o600))

	cfg, err := Load(filename)
	require.NoError(t, err)
	assert.Empty(t, cfg.Defaults.WatchMode)
	require.NotNil(t, cfg.Defaults.WatchRepo)
	assert.True(t, *cfg.Defaults.WatchRepo)

	image := model.Image{}
	require.NoError(t, image.ResolveWatchMode(cfg.Defaults))
	assert.Equal(t, registry.WatchModeAll, image.WatchMode)
}

func TestWatchModeConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		env     string
		want    registry.WatchMode
		wantErr bool
	}{
		{
			name: "file mode",
			mode: "semver",
			want: registry.WatchModeSemver,
		},
		{
			name: "environment overrides file",
			mode: "all",
			env:  "semver+prereleases",
			want: registry.WatchModeSemverPrereleases,
		},
		{
			name:    "invalid file mode",
			mode:    "invalid",
			wantErr: true,
		},
		{
			name:    "invalid environment mode",
			mode:    "all",
			env:     "invalid",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "diun.yml")
			data := "defaults:\n  watchMode: " + tt.mode + "\n  watchRepo: true\nproviders:\n  file:\n    filename: ./fixtures/file.yml\n"
			require.NoError(t, os.WriteFile(filename, []byte(data), 0o600))
			if tt.env != "" {
				t.Setenv("DIUN_DEFAULTS_WATCHMODE", tt.env)
			}
			cfg, err := Load(filename)
			if tt.wantErr {
				require.ErrorContains(t, err, "WatchMode")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Defaults.WatchMode)
			image := model.Image{}
			require.NoError(t, image.ResolveWatchMode(cfg.Defaults))
			assert.Equal(t, tt.want, image.WatchMode)
		})
	}
}

func TestValidateMetricsMetadataLabels(t *testing.T) {
	cases := []struct {
		name    string
		labels  []string
		wantErr string
	}{
		{
			name: "Default labels",
		},
		{
			name:   "Valid labels",
			labels: []string{"channel", "environment"},
		},
		{
			name:    "Duplicate labels",
			labels:  []string{"channel", "channel"},
			wantErr: "'MetadataLabels' failed on the 'unique' tag",
		},
		{
			name:    "Empty label",
			labels:  []string{""},
			wantErr: `invalid metrics metadata label key ""`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Db:        &model.Db{Path: filepath.Join(t.TempDir(), "diun.db")},
				Watch:     (&model.Watch{}).GetDefaults(),
				Metrics:   (&model.Metrics{}).GetDefaults(),
				Providers: &model.Providers{},
			}
			cfg.Metrics.MetadataLabels = tt.labels

			err := cfg.validate()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidation(t *testing.T) {
	cases := []struct {
		name string
		cfg  string
	}{
		{
			name: "Success",
			cfg:  "./fixtures/config.validate.yml",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.cfg)
			require.NoError(t, err)
			_, err = env.Encode("DIUN_", cfg)
			require.NoError(t, err)
		})
	}
}
