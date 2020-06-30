## What is Diun?

**D**ocker **I**mage **U**pdate **N**otifier is a CLI application written in [Go](https://golang.org/) and delivered as a
[single executable](https://github.com/crazy-max/diun/releases/latest) (and a [Docker image](install/docker.md))
to receive notifications when a Docker image is updated on a Docker registry.

With Go, this can be done with an independent binary distribution across all platforms and architectures that Go supports.
This support includes Linux, macOS, and Windows, on architectures like amd64, i386, ARM, PowerPC, and others.

## Features

* Allow to watch a Docker repository and report new tags
* Include and exclude filters with regular expression for tags
* Internal cron implementation through go routines
* Worker pool to parallelize analyses
* Allow overriding image os and architecture
* [Docker](providers/docker.md), [Swarm](providers/swarm.md), [Kubernetes](providers/kubernetes.md)
and [File](providers/file.md) providers available
* Get notified through Gotify, Mail, Slack, Telegram and [more](config/index.md#reference)
* Enhanced logging
* Timezone can be changed
* Official [Docker image available](install/docker.md)

## Usage

* [Command line](usage/cli.md)
* [Basic example](usage/basic-example.md)
* [Configuration](config/index.md)
