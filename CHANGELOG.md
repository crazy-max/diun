# Changelog

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
* Update libs
* Go 1.13.5
* Seconds field optional for schedule

> :warning: **BREAKING CHANGES**<br />
> `image` field has been moved to `providers.static` in configuration file.<br />
> See [providers configuration](doc/configuration.md#providers) for more info.

## 1.4.1 (2019/10/20)

* Update libs
* Fix Docker labels

## 1.4.0 (2019/10/01)

* Multi-platform Docker image
* Switch to GitHub Actions
* :warning: Run Docker container as non-root user
* :warning: Stop publishing Docker image on Quay
* Go 1.12.10
* Use GOPROXY

> :warning: **UPGRADE NOTES**
> As the Docker container now runs as a non-root user, you have to first stop the container and change permissions to `data` volume:
> ```
> docker-compose stop
> chown -R 1000:1000 data/
> docker-compose pull
> docker-compose up -d
> ```

## 1.3.0 (2019/08/22)

* Add Linux service doc and sample
* Move documentation
* Fix go mod
* Remove `--docker` flag
* Allow to override database path through `DIUN_DB` env var

## 1.2.0 (2019/08/18)

* Update libs
* Display containers/image logs
* Fix registry options not setted (Issue #5)

## 1.1.0 (2019/07/24)

* Update libs

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

> :warning: **BREAKING CHANGES**
> Some fields in configuration file has been changed:
> * `registries` renamed `regopts`
> * `items` renamed `image`
> * `items[].image` renamed `image[].name`
> * `items[].registry_id` renamed `image[].regopts_id`
> * `watch.os` and `watch.arch` moved to `image[].os` and `image[].arch`
> See README for more info.

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
