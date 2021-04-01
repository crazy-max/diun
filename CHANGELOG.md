# Changelog

## 4.14.0 (2021/03/15)

* Bump k8s.io/client-go from 0.19.4 to 0.20.4 (#280)
* Docker client 20.10.5 (#303)
* Allow telegram chat IDs as file (#301)
* Go 1.16 (#302)
* Handle git ref for artifact target
* Bump github.com/imdario/mergo from 0.3.11 to 0.3.12 (#298)
* Bump github.com/crazy-max/gohealthchecks from 0.2.0 to 0.3.0 (#296)
* Bump github.com/alecthomas/kong from 0.2.15 to 0.2.16 (#295)
* Allow to configure scheme for MQTT broker (#292)
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

* Allow to set the hostname sent to the SMTP server with the HELO command for mail notification (#165)
* Fix Telegram notification error (#162)

## 4.4.1 (2020/08/20)

* Allow to use `--test-notif` without providers and DB connection (#157 #150)
* Update deps

## 4.4.0 (2020/08/08)

* Allow to customize message type for Matrix notifications (#143)

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
* Allow to override database path through `DIUN_DB` env var

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
