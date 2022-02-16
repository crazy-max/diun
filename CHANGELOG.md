# Changelog

## 4.21.0 (2022/01/26)

* Add `image prune` command (#519)
* Fix matrix login scheme (#487)
* Move `syscall` to `golang.org/x/sys` (#525)
* Move from `io/ioutil` to `os` package (#524)
* Fix notif template in docs
* Enhance dockerfiles (#523)
* Add binary bake target (#517)
* MkDocs Material 8.1.8 (#520 #548)
* Alpine Linux 3.15 (#527)
* goreleaser-xx 1.2.5 (#539)
* Bump github.com/alecthomas/kong from 0.2.17 to 0.3.0 (#507 #537)
* Bump github.com/containerd/containerd from 1.5.5 to 1.5.8 (#494 #496 #509)
* Bump github.com/containers/image/v5 from 5.16.0 to 5.19.0 (#498 #536 #546)
* Bump github.com/docker/docker from 20.10.8+incompatible to 20.10.12+incompatible (#500 #510 #531)
* Bump github.com/go-playground/validator/v10 from 10.9.0 to 10.10.0 (#538)
* Bump github.com/jedib0t/go-pretty/v6 from 6.2.4 to 6.2.5 (#543)
* Bump github.com/microcosm-cc/bluemonday from 1.0.15 to 1.0.17 (#499 #535)
* Bump github.com/moby/buildkit from 0.9.0 to 0.9.3 (#495 #506 #512)
* Bump github.com/opencontainers/image-spec to v1.0.2-0.20211117181255-693428a734f5 (#513)
* Bump github.com/panjf2000/ants/v2 from 2.4.6 to 2.4.7 (#532)
* Bump github.com/rs/zerolog from 1.24.0 to 1.26.1 (#485 #502 #534)
* Bump google.golang.org/grpc from 1.40.0 to 1.44.0 (#492 #505 #529 #545)
* Bump google.golang.org/grpc/cmd/protoc-gen-go-grpc from 1.1.0 to 1.2.0 (#533)
* Bump k8s.io/client-go from 0.22.1 to 0.22.4 (#490 #511)

## 4.20.1 (2021/09/06)

* Fix notification title (#483)

## 4.20.0 (2021/09/05)

* Option to render fields (#480)
* Allow choosing status to be notified (#475)
* Enhance notif wording (#467)
* Wrong remaining time displayed (#469)
* Allow multi recipients for email notifier (#463)
* Provide mutable tags for Diun image (#462)
* Fix Dockerfile parser and add tests (#459)
* Add e2e tests (#471)
* Use args in kubernetes documentation example (#424)
* Fix j2 variable in docs (#422)
* Note to customize the hostname (#465)
* Go 1.17 (#458)
* Add `windows/arm64` artifact (#472)
* Add `linux/riscv64` artifact (#427)
* Alpine Linux 3.14 (#426)
* MkDocs Material 7.2.6 (#428 #482)
* Protoc 3.17.3 (#461)
* Bump github.com/rs/zerolog from 1.23.0 to 1.24.0 (#477)
* Bump github.com/crazy-max/gonfig from 0.4.0 to 0.5.0 (#474)
* Bump github.com/gregdel/pushover to 1.1.0 (#470)
* Bump github.com/streadway/amqp to 1.0.0 (#470)
* Bump github.com/containers/image/v5 from 5.13.2 to 5.16.0 (#460 #476)
* Bump k8s.io/client-go from 0.21.2 to 0.22.1 (#466)
* Bump github.com/docker/docker from 20.10.7 to 20.10.8 (#451)
* Bump github.com/moby/buildkit from 0.8.3 to 0.9.0 (#437)
* Bump github.com/containerd/containerd from 1.5.2 to 1.5.5 (#433 #440 #447)
* Bump github.com/microcosm-cc/bluemonday from 1.0.14 to 1.0.15 (#430)
* Bump github.com/go-playground/validator/v10 from 10.6.1 to 10.9.0 (#429 #445 #455)
* Bump github.com/jedib0t/go-pretty/v6 from 6.2.2 to 6.2.4 (#432)
* Bump google.golang.org/grpc from 1.38.0 to 1.40.0 (#421 #453 #456)
* Bump google.golang.org/protobuf from 1.26.0 to 1.27.1 (#420)
* Bump codecov/codecov-action from 1 to 2

## 4.19.0 (2021/06/26)

* Allow customizing notification message (#415)
* Bump github.com/panjf2000/ants/v2 from 2.4.5 to 2.4.6 (#416)
* Bump k8s.io/client-go from 0.21.1 to 0.21.2 (#414)
* Bump github.com/microcosm-cc/bluemonday from 1.0.13 to 1.0.14 (#413)
* Bump github.com/containers/image/v5 from 5.13.1 to 5.13.2 (#412)

## 4.18.0 (2021/06/18)

* Handle registry auth config (#411)
* Bump k8s.io/client-go from 0.20.6 to 0.21.1 (#381)
* Bump github.com/containers/image/v5 from 5.12.0 to 5.13.1 (#409)
* Avoid notification for unupdated image (#406)
* Use `openssl` pkg (#407)
* Bump github.com/rs/zerolog from 1.22.0 to 1.23.0 (#405)
* Bump google.golang.org/grpc from 1.37.0 to 1.38.0 (#389)
* Bump github.com/microcosm-cc/bluemonday from 1.0.9 to 1.0.13 (#403 #410)
* Bumps github.com/docker/docker from 20.10.6+incompatible to 20.10.7+incompatible (#397)
* Bump github.com/jedib0t/go-pretty/v6 from 6.2.1 to 6.2.2 (#388)
* Bump go.etcd.io/bbolt from 1.3.5 to 1.3.6 (#394)
* Bump github.com/eclipse/paho.mqtt.golang from 1.3.4 to 1.3.5 (#400)
* Bump github.com/alecthomas/kong from 0.2.16 to 0.2.17 (#401)
* Bump github.com/tidwall/pretty from 1.1.0 to 1.2.0 (#390 #402)
* Set `cacheonly` output for validators (#395)
* Define serve command (#393)
* Save raw manifest in db (#391)

## 4.17.0 (2021/05/26)

:warning: See **Migration notes** in the documentation before upgrading.

* Add CLI to interact with Diun through gRPC (#382)
    * Create `image` and `notif` proto services
    * Implement proto definitions
    * New commands `serve`, `image` and `notif`
    * Refactor command line usage doc
    * Better CLI error handling
    * Tools build constraint to manage tools deps through go modules
    * Compile and validate protos through a dedicated Dockerfile and a bake target    
    * Merge validate and build workflow
    * Add upgrade notes
* Bump github.com/eclipse/paho.mqtt.golang from 1.3.3 to 1.3.4 (#359)
* Bump github.com/panjf2000/ants/v2 from 2.4.4 to 2.4.5 (#380)
* Bump github.com/rs/zerolog from 1.21.0 to 1.22.0 (#379)
* Bump github.com/go-playground/validator/v10 from 10.5.0 to 10.6.1 (#377)
* MkDocs Materials 7.1.5 (#386)
* Add `NO_COLOR` support (#384)
* Bump github.com/pkg/profile from 1.5.0 to 1.6.0 (#363)
* Move to docker/metadata-action (#366)
* Bump github.com/containers/image/v5 from 5.11.1 to 5.12.0 (#360)
* Bump github.com/containerd/containerd from 1.5.0-beta.4 to 1.5.2 (#353 #361 #362 #383)
* Add blog posts (#355 #385)
* Bump github.com/moby/buildkit from 0.8.2 to 0.8.3 (#354)

## 4.16.1 (2021/04/30)

* Fix Swarm Provider (#351)

## 4.16.0 (2021/04/29)

* Dockerfile provider (#329)
* Note about `watch_repo` setting (#348)
* Contribute to doc (#347)
* Update docs for Podman support (#345)
* Optional profiler volume (#344)

## 4.15.2 (2021/04/25)

* Make profiler optional (#341)

## 4.15.1 (2021/04/25)

* Fix profiler path (#339)

## 4.15.0 (2021/04/25)

* Add `darwin/arm64` artifact (#338)
* MkDocs Materials 7.1.3 (#337)
* Add profiler flag (#336)
* Handle digest based image reference (#335)
* Bump github.com/docker/docker from 20.10.5+incompatible to 20.10.6+incompatible (#324)
* Bump github.com/containers/image/v5 from 5.10.5 to 5.11.1 (#323 #330)
* Bump github.com/go-playground/validator/v10 from 10.4.1 to 10.5.0 (#319)
* Bump github.com/panjf2000/ants/v2 from 2.4.3 to 2.4.4 (#312)
* Bump github.com/rs/zerolog from 1.20.0 to 1.21.0 (#309)
* Bump github.com/microcosm-cc/bluemonday from 1.0.4 to 1.0.9 (#311 #321 #325 #333)
* Bump github.com/eclipse/paho.mqtt.golang from 1.3.2 to 1.3.3 (#316)
* Deploy docs on workflow dispatch or tag (#305)

## 4.14.0 (2021/03/15)

* Bump k8s.io/client-go from 0.19.4 to 0.20.4 (#280)
* Docker client 20.10.5 (#303)
* Allow telegram chat IDs as file (#301)
* Go 1.16 (#302)
* Handle git ref for artifact target
* Bump github.com/imdario/mergo from 0.3.11 to 0.3.12 (#298)
* Bump github.com/crazy-max/gohealthchecks from 0.2.0 to 0.3.0 (#296)
* Bump github.com/alecthomas/kong from 0.2.15 to 0.2.16 (#295)
* Allow configuring scheme for MQTT broker (#292)
* Switch to [goreleaser-xx](https://github.com/crazy-max/goreleaser-xx) (#291)
* Bump github.com/containers/image/v5 from 5.10.4 to 5.10.5 (#290)

## 4.13.0 (2021/03/01)

* Missing token as secret setting for some notifiers (#289)
* Allow disabling log color output (#288)
* Bump github.com/containers/image/v5 from 5.10.1 to 5.10.4 (#271 #282 #284)
* Cleanup workflows (#281 #287)
* Do not check recipient details for Pushover (#277)
* MkDocs Materials 6.2.8 (#276)
* Fix markdown renderer (#275)
* Add message client for notifiers (#273)

## 4.12.0 (2021/02/09)

* Use digest as comparison footprint (#269)
* Bump github.com/alecthomas/kong from 0.2.12 to 0.2.15 (#270)
* Bump github.com/eclipse/paho.mqtt.golang from 1.3.1 to 1.3.2 (#268)
* Bump github.com/containers/image/v5 from 5.9.0 to 5.10.1 (#265)
* Move to [docker/bake-action](https://github.com/docker/bake-action) (#266)
* Typo in documentation (#258)
* Log image validation

## 4.11.0 (2021/01/04)

* Fix DB migration (#255)
* Add Pushover notification (#254)
* Avoid duplicated notifications with Kubernetes DaemonSet (#252)
* Make scheduler optional (#251)
* Bump github.com/eclipse/paho.mqtt.golang from 1.3.0 to 1.3.1 (#249)
* Handle exclusions as a distinct status (#248)

## 4.10.0 (2020/12/26)

* Refactor CI and dev workflow with buildx bake (#247)
    * Upload artifacts
    * Add `image-local` target
    * Single job for artifacts and image
    * Add `armv5` artifact
* MQTT Reconnection Log Spam (#241)
* Add Docker + File providers user guide (#239)
* Bump github.com/alecthomas/kong from 0.2.11 to 0.2.12 (#231)
* Bump github.com/eclipse/paho.mqtt.golang from 1.2.0 to 1.3.0 (#235)
* Bump github.com/containers/image/v5 from 5.8.1 to 5.9.0 (#236)
* Bump gopkg.in/yaml.v2 from 2.3.0 to 2.4.0 (#228)
* Bump github.com/containers/image/v5 from 5.8.0 to 5.8.1 (#226)

## 4.9.0 (2020/11/16)

* Fix duplicated notifications
* Remove support for `freebsd/*` (moby/moby#38818)
* Add support for `linux/ppc64le` and `linux/s390x` (binary)
* Bump k8s.io/client-go from 0.19.3 to 0.19.4 (#224)
* Bump github.com/containers/image/v5 to 5.8.0

## 4.8.1 (2020/11/14)

* Fix registry timeout context (#221)
* Image closer not required while fetching tags

## 4.8.0 (2020/11/13)

* Go 1.15 (#218)
* Remove `linux/s390x` platform support (for now)
* Check digest from HEAD request (#217)
* Add FAQ note about Docker Hub rate limits
* Compare digest as watch setting
* Optimize build time
* Add hub link for GitHub Container Registry (#211)
* Update deps

## 4.7.0 (2020/11/02)

* Add MQTT notification (#192)
* Docker image also available on [GitHub Container Registry](https://github.com/users/crazy-max/packages/container/package/diun)
* Use zoneinfo from Go (#202)
* Remove `--timezone` flag
* Use Docker meta action to handle tags and labels
* Update deps

## 4.6.1 (2020/10/22)

* Typos in documentation
* Bump docker/login-action from v1.4.1 to v1.6.0 (#198)
* Bump k8s.io/client-go from 0.19.2 to 0.19.3 (#199)
* Bump codecov/codecov-action from v1.0.13 to v1.0.14 (#195)
* Bump github.com/go-playground/validator/v10 from 10.4.0 to 10.4.1 (#197)
* Bump github.com/panjf2000/ants/v2 from 2.4.2 to 2.4.3 (#196)

## 4.6.0 (2020/10/13)

* Add support for [Healthchecks](https://healthchecks.io/) to monitor Diun watcher (#78)
* Add option to mention specific users or roles for Discord notifier (#188)
* Update docker install documentation
* Add "Too many requests to registry" section in FAQ (#168)
* Update deps
* Switch to [Docker actions](https://github.com/docker/build-push-action)

## 4.5.0 (2020/08/29)

* Allow setting the hostname sent to the SMTP server with the HELO command for mail notification (#165)
* Fix Telegram notification error (#162)

## 4.4.1 (2020/08/20)

* Allow using `--test-notif` without providers and DB connection (#157 #150)
* Update deps

## 4.4.0 (2020/08/08)

* Allow customizing message type for Matrix notifications (#143)

## 4.3.1 (2020/07/30)

* Hostname not taken into account for Matrix notifications

## 4.3.0 (2020/07/29)

* Add Matrix notification (#124)

## 4.2.0 (2020/07/16)

* Seek configuration file from default places (#107)
* Switch to [gonfig](https://github.com/crazy-max/gonfig)
* Update deps

## 4.1.1 (2020/06/26)

* Small typo

## 4.1.0 (2020/06/26)

* Discord notifications (#110 #111)
* Update migration notes (#107)
* Logging when configuration file or `DIUN_` env vars not found (#107)
* Bump github.com/containers/image/v5 from 5.4.4 to 5.5.1 (#96)

## 4.0.0 (2020/06/22)

:warning: See **Migration notes** in the documentation for breaking changes.

* Display hostname in notifications (#102)
* Automatically determine registry options based on image name (#103)
* Docs website with mkdocs (#99)
* Skip dangling images (#98)
* More explicit message if manifest not found (#94)
* Add swarm example
* Update doc for file and Swarm providers
* Add Kubernetes provider (#25)
* Update Teams notification screenshot (#93)
* Send message as markdown for Gotify and Telegram notifiers
* Add link to respective hub (#40)
* Configuration transposed into environment variables (#82)
* Configuration file not required anymore
* `DIUN_DB` env var renamed `DIUN_DB_PATH`
* Only accept duration as timeout value (`10` becomes `10s`)
* Enhanced documentation (#83)
* Add note about test notifications (#79)
* Improve configuration validation
* Fix telegram init
* All fields in configuration are now _camelCased_
* Docker API version negotiation (#29)
* Add Mattermost compatibility via Slack webhooks (#80)
* Update deps

## 3.0.0 (2020/05/27)

:warning: See **Migration notes** in the documentation for breaking changes.

* Add script notification (#53)
* Add Teams notification (#72)
* Add `--test-notif` flag (#23)
* Allow only one Docker and Swarm provider
* Remove "enable" setting for notifiers
* Logging when no image is found
* Add Amqp notification client (#63)
* Fix default log level
* Move static to file provider (#71)
* Reload config on change for file provider (#16)
* Switch to kong command-line parser (#66)
* Enhanced Dockerfile
* Review of platform detection (#57)
* Leave default image platform empty for file provider (see FAQ doc)
* Handle platform variant
* Add database migration process
* Switch to Open Container Specification labels as label-schema.org ones are deprecated
* Remove unneeded `diun.os` and `diun.arch` docker labels
* Add upgrade notes
* Update deps

## 2.6.1 (2020/03/26)

* Downgrade containers/image to 5.2.1 (#54)

## 2.6.0 (2020/03/26)

* Fix service image inspection (#52)
* Docker client v19.03.8
* Update deps

## 2.5.0 (2020/03/01)

* Add Rocket.Chat notifier (#44)

## 2.4.0 (2020/02/17)

* Add Gotify notification client (#36)
* Bump containers/image v5 (#35) 

## 2.3.0 (2020/01/28)

* Add Telegram notifier (#30)
* Docker client struct options
* Move registry client to a dedicated package

## 2.2.1 (2020/01/07)

* Set user agent for Docker registry client
* Update deps

## 2.2.0 (2019/12/22)

* Add option to skip notification at the very first analysis of an image (#10)
* Skip analysis of locally built image

## 2.1.0 (2019/12/17)

* Add Slack notifier (#8)

## 2.0.0 (2019/12/14)

:warning: See **Migration notes** in the documentation for breaking changes.

* Include provider in notifications
* Add providers documentation
* Move image validation and improve job execution
* Add Swarm provider
* Add fields to load sensitive values from file (#7)
* Add Docker provider (#3)
* Docker client v19.03.5
* Move `image` field to providers layer and rename it `static`
* Update deps
* Go 1.13.5
* Seconds field optional for schedule

## 1.4.1 (2019/10/20)

* Update deps
* Fix Docker labels

## 1.4.0 (2019/10/01)

* Multi-platform Docker image
* Switch to GitHub Actions
* Stop publishing Docker image on Quay
* Go 1.12.10
* Use GOPROXY

## 1.3.0 (2019/08/22)

* Add Linux service doc and sample
* Move documentation
* Fix go mod
* Remove `--docker` flag
* Allow overriding database path through `DIUN_DB` env var

## 1.2.0 (2019/08/18)

* Update deps
* Display containers/image logs
* Fix registry options not setted (#5)

## 1.1.0 (2019/07/24)

* Update deps

## 1.0.2 (2019/07/01)

* Worker pool can be full while retrieving tags

## 1.0.1 (2019/07/01)

* Fix runtime error

## 1.0.0 (2019/07/01)

:warning: See **Migration notes** in the documentation for breaking changes.

* Always run on startup. Flag `--run-startup` removed.
* Display next execution time
* Use v3 robfig/cron
* Move `Os` and `Arch` filters to image
* Retrieve all tags by default
* Review config file structure
* Improve worker pool

## 0.5.0 (2019/06/09)

* Add worker pool to parallelize analyses

## 0.4.1 (2019/06/08)

* Filter tags before return them

## 0.4.0 (2019/06/08)

* Add option to set the maximum number of tags to watch for an item if `watch_repo` is enabled

## 0.3.2 (2019/06/08)

* Fix registry client context

## 0.3.1 (2019/06/08)

* Fix email template
* Add flag to log caller

## 0.3.0 (2019/06/08)

* Allow overriding os and architecture when watching
* Move `insecure_tls` and `timeout` options to registry option
* Rename Bolt bucket
* Change default schedule
* Review registry client

## 0.2.0 (2019/06/05)

* Don't skip repo analysis if default tag not found
* Docker engine 18.09.6

## 0.1.1 (2019/06/04)

* Increase default timeout
* Fix `data` volume mount

## 0.1.0 (2019/06/04)

* Initial version
