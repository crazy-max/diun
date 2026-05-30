<img src="assets/logo.png" alt="Diun" width="128px" style="display: block; margin-left: auto; margin-right: auto"/>

<p align="center">
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/diun.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/diun/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/diun/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/diun/actions?workflow=build"><img src="https://img.shields.io/github/actions/workflow/status/crazy-max/diun/build.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/stars/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/diun/"><img src="https://img.shields.io/docker/pulls/crazymax/diun.svg?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/diun"><img src="https://goreportcard.com/badge/github.com/crazy-max/diun?style=flat-square" alt="Go Report"></a>
  <a href="https://codecov.io/gh/crazy-max/diun"><img src="https://img.shields.io/codecov/c/github/crazy-max/diun?logo=codecov&style=flat-square" alt="Codecov"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

---

## What is Diun?

**D**ocker **I**mage **U**pdate **N**otifier helps you keep track of container
image updates without manually watching registries.

Diun checks your images on a schedule, detects when a tracked tag or digest has
changed, and notifies you when something new is available. That makes it useful
for staying on top of upstream base image rebuilds, application releases, and
other image changes that can otherwise slip by unnoticed.

You can run Diun as a [single executable](https://github.com/crazy-max/diun/releases/latest)
or as a [Docker image](install/docker.md), point it at the container sources you
care about, and send notifications to the messaging services your team already
uses.

## Features

* Watch container images and report when tags or digests change
* Track repositories with include and exclude filters for tags
* Run checks on a schedule without needing an external cron job
* Discover images from [Docker](providers/docker.md), [Containerd](providers/containerd.md),
  [Kubernetes](providers/kubernetes.md), [Swarm](providers/swarm.md), [Nomad](providers/nomad.md),
  [Dockerfile](providers/dockerfile.md), and [File](providers/file.md) providers
* Override target image OS and architecture when needed
* Send notifications through Gotify, Mail, Slack, Telegram, and [more](config/index.md#reference)
* Integrate with [Healthchecks](config/watch.md#healthchecks) to monitor the watcher itself
* Use detailed logging to understand checks, updates, and delivery status
* Run as a [single binary](https://github.com/crazy-max/diun/releases/latest) or with the official
  [Docker image](install/docker.md)

## License

This project is licensed under the terms of the MIT license.
