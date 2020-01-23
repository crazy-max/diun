module github.com/crazy-max/diun

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/alecthomas/kingpin v0.0.0-20190816080609-dce89ec0b9f1
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/containerd/continuity v0.0.0-20191127005431-f65d91d395eb // indirect
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/storage v1.15.3 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hako/durafmt v0.0.0-20190612201238-650ed9f29a84
	github.com/imdario/mergo v0.3.8
	github.com/matcornic/hermes/v2 v2.0.2
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nlopes/slack v0.6.0
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/panjf2000/ants/v2 v2.2.2
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.17.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	go.etcd.io/bbolt v1.3.3
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
