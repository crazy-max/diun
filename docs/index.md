<img src="assets/logo.png" alt="Diun" width="128px" style="display: block; margin-left: auto; margin-right: auto"/>

<p align="center">
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/diun.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/diun/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/diun/actions?workflow=build"><img src="https://img.shields.io/github/workflow/status/crazy-max/diun/build?label=build&logo=github&style=flat-square" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/stars/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/pulls/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/diun"><img src="https://goreportcard.com/badge/github.com/crazy-max/diun?style=flat-square" alt="Go Report"></a>
  <a href="https://codecov.io/gh/crazy-max/diun"><img src="https://img.shields.io/codecov/c/github/crazy-max/diun?logo=codecov&style=flat-square" alt="Codecov"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

---

## What is Diun?

**D**ocker **I**mage **U**pdate **N**otifier is a CLI application written in [Go](https://golang.org/) and delivered as a
[single executable](https://github.com/crazy-max/diun/releases/latest) (and a [Docker image](install/docker.md))
to receive notifications when a Docker image is updated on a Docker registry.

With Go, this can be done with an independent binary distribution across all platforms and architectures that Go supports.
This support includes Linux, macOS, and Windows, on architectures like amd64, i386, ARM, PowerPC, and others.

## Features

* Allow watching a Docker repository and report new tags
* Include and exclude filters with regular expression for tags
* Internal cron implementation through go routines
* Worker pool to parallelize analyses
* Allow overriding image os and architecture
* [Docker](providers/docker.md), [Swarm](providers/swarm.md), [Kubernetes](providers/kubernetes.md),
[Dockerfile](providers/dockerfile.md) and [File](providers/file.md) providers available
* Get notified through Gotify, Mail, Slack, Telegram and [more](config/index.md#reference)
* [Healthchecks support](config/watch.md#healthchecks) to monitor Diun watcher
* Enhanced logging
* Official [Docker image available](install/docker.md)

## License

This project is licensed under the terms of the MIT license.
