module github.com/crazy-max/diun/v4

go 1.15

require (
	github.com/alecthomas/kong v0.2.16
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/containerd/containerd v1.5.0
	github.com/containers/image/v5 v5.12.0
	github.com/crazy-max/gohealthchecks v0.3.0
	github.com/crazy-max/gonfig v0.4.0
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/eclipse/paho.mqtt.golang v1.3.3
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-playground/validator/v10 v10.5.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gregdel/pushover v0.0.0-20201104094836-ddbe0c1d3a38
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.12
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/matrix-org/gomatrix v0.0.0-20200501121722-e5578b12c752
	github.com/microcosm-cc/bluemonday v1.0.9
	github.com/moby/buildkit v0.8.3
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/nlopes/slack v0.6.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/panjf2000/ants/v2 v2.4.4
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.21.0
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.7.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.etcd.io/bbolt v1.3.5
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

// containerd: corresponds to buildkit
replace github.com/containerd/containerd => github.com/containerd/containerd v1.4.1-0.20201117152358-0edc412565dc
