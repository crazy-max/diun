module github.com/crazy-max/diun

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/alecthomas/kingpin v0.0.0-20190816080609-dce89ec0b9f1
	github.com/containers/image/v5 v5.2.1
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.8
	github.com/matcornic/hermes/v2 v2.0.2
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nlopes/slack v0.6.0
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/panjf2000/ants/v2 v2.2.2
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.17.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.etcd.io/bbolt v1.3.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
