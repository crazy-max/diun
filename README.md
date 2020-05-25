<p align="center"><a href="https://github.com/crazy-max/diun" target="_blank"><img height="128" src="https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png"></a></p>

<p align="center">
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/diun.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/diun/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/diun/actions"><img src="https://github.com/crazy-max/diun/workflows/build/badge.svg" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/stars/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/pulls/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/diun"><img src="https://goreportcard.com/badge/github.com/crazy-max/diun?style=flat-square" alt="Go Report"></a>
  <a href="https://www.codacy.com/app/crazy-max/diun"><img src="https://img.shields.io/codacy/grade/f2ef980c87d247ce8a8dbc98a8f4f434.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**Diun** is a CLI application written in [Go](https://golang.org/) to receive notifications when a Docker image is updated on a Docker registry. With Go, this app can be used across many platforms and architectures. This support includes Linux, FreeBSD, macOS and Windows on architectures like amd64, i386, ARM and others.

![](.res/notif-slack.png)

## Features

* Allow to watch a Docker repository and report new tags
* Include and exclude filters with regular expression for tags
* Internal cron implementation through go routines
* Worker pool to parallelize analyses
* Allow overriding image os and architecture
* Multi providers available like [Docker](doc/providers/docker.md), [Swarm](doc/providers/swarm.md), [File](doc/providers/file.md)...
* Get notified through Slack, Mail, Telegram and [more](doc/notifications.md)
* Enhanced logging
* Timezone can be changed
* Official [Docker image available](doc/install/docker.md)

## Documentation

* Install
  * [With Docker](doc/install/docker.md)
  * [From binary](doc/install/binary.md)
  * [Linux service](doc/install/linux-service.md)
* [Usage](doc/usage.md)
* [Configuration](doc/configuration.md)
* Providers
  * [Docker](doc/providers/docker.md)
  * [Swarm](doc/providers/swarm.md)
  * [File](doc/providers/file.md)
* [Notifications](doc/notifications.md)
* [FAQ](doc/faq.md)

## How can I help?

All kinds of contributions are welcome :raised_hands:! The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon: You can also support this project by [**becoming a sponsor on GitHub**](https://github.com/sponsors/crazy-max) :clap: or by making a [Paypal donation](https://www.paypal.me/crazyws) to ensure this journey continues indefinitely! :rocket:

Thanks again for your support, it is much appreciated! :pray:

## License

MIT. See `LICENSE` for more details.
