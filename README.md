<p align="center"><a href="https://github.com/crazy-max/diun" target="_blank"><img height="128" src="https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png"></a></p>

<p align="center">
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/diun.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/diun/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/diun/actions"><img src="https://github.com/crazy-max/diun/workflows/build/badge.svg" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/stars/crazymax/diun.svg?style=flat-square" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/pulls/crazymax/diun.svg?style=flat-square" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/diun"><img src="https://goreportcard.com/badge/github.com/crazy-max/diun?style=flat-square" alt="Go Report"></a>
  <a href="https://www.codacy.com/app/crazy-max/diun"><img src="https://img.shields.io/codacy/grade/f2ef980c87d247ce8a8dbc98a8f4f434.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://www.patreon.com/crazymax"><img src="https://img.shields.io/badge/donate-patreon-f96854.svg?logo=patreon&style=flat-square" alt="Support me on Patreon"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**Diun** :bell: is a CLI application written in [Go](https://golang.org/) to  receive notifications :inbox_tray: when a Docker :whale: image is updated on a Docker registry. With Go, this app can be used across many platforms :game_die: and architectures. This support includes Linux, FreeBSD, macOS and Windows on architectures like amd64, i386, ARM and others.

## Features

* Allow to watch a full Docker repository and report new tags
* Include and exclude filters with regular expression for tags
* Internal cron implementation through go routines
* Worker pool to parallelize analyses
* Allow overriding image os and architecture
* Beautiful email report
* Webhook notification
* Enhanced logging
* Timezone can be changed
* :whale: Official [Docker image available](doc/install/docker.md)

## Documentation

* Install
  * [With Docker](doc/install/docker.md)
  * [From binary](doc/install/binary.md)
  * [Linux service](doc/install/linux-service.md)
* [Usage](doc/usage.md)
* [Configuration](doc/configuration.md)
* [Notifications](doc/notifications.md)
* [TODO](doc/todo.md)

## How can I help ?

All kinds of contributions are welcome :raised_hands:!<br />
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:<br />
But we're not gonna lie to each other, I'd rather you buy me a beer or two :beers:!

[![Support me on Patreon](.res/patreon.png)](https://www.patreon.com/crazymax) 
[![Paypal](.res/paypal.png)](https://www.paypal.me/crazyws)

## License

MIT. See `LICENSE` for more details.
