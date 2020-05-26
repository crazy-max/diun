# Changelog

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

> :warning: See [**UPGRADE NOTES**](UPGRADE.md#1x--2x) for breaking changes.

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
* Fix registry options not setted (Issue #5)

## 1.1.0 (2019/07/24)

* Update deps

## 1.0.2 (2019/07/01)

* Worker pool can be full while retrieving tags

## 1.0.1 (2019/07/01)

* Fix runtime error

## 1.0.0 (2019/07/01)

* Always run on startup. Flag `--run-startup` removed.
* Display next execution time
* Use v3 robfig/cron
* Move `Os` and `Arch` filters to image
* Retrieve all tags by default
* Review config file structure
* Improve worker pool

> :warning: See [**UPGRADE NOTES**](UPGRADE.md#0x--1x) for breaking changes.

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
