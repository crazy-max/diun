module github.com/crazy-max/diun

require (
	github.com/alecthomas/kingpin v0.0.0-20190816080609-dce89ec0b9f1
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/storage v1.13.4 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.8
	github.com/matcornic/hermes/v2 v2.0.2
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/panjf2000/ants v1.2.0
	github.com/robfig/cron/v3 v3.0.0
	github.com/rs/zerolog v1.16.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	go.etcd.io/bbolt v1.3.3
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190423201726-d2cfbce3f3b0
