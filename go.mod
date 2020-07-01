module github.com/crazy-max/diun/v4

go 1.13

require (
	github.com/alecthomas/kong v0.2.9
	github.com/containers/image/v5 v5.5.1
	github.com/containous/traefik/v2 v2.2.1
	github.com/docker/docker v1.4.2-0.20200204220554-5f6d6f3f2203
	github.com/docker/go-connections v0.4.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-playground/validator/v10 v10.3.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.9
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/nlopes/slack v0.6.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/panjf2000/ants/v2 v2.4.1
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.19.0
	github.com/sirupsen/logrus v1.6.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.6.1
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.etcd.io/bbolt v1.3.5
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

// Containous forks
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
)
