module github.com/crazy-max/diun/v4

go 1.13

require (
	github.com/alecthomas/kong v0.2.11
	github.com/containers/image/v5 v5.6.0
	github.com/crazy-max/gonfig v0.3.0
	github.com/docker/docker v1.4.2-0.20200204220554-5f6d6f3f2203
	github.com/docker/go-connections v0.4.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-playground/validator/v10 v10.4.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.11
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/matrix-org/gomatrix v0.0.0-20200501121722-e5578b12c752
	github.com/microcosm-cc/bluemonday v1.0.4
	github.com/nlopes/slack v0.6.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/panjf2000/ants/v2 v2.4.2
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.20.0
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.6.1
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	go.etcd.io/bbolt v1.3.5
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203
