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

## Diun CLI

```
$ ./diun --help
Usage: diun

Docker image update notifier. More info: https://github.com/crazy-max/diun

Flags:
  --help                Show context-sensitive help.
  --version
  --config=STRING       Diun configuration file ($CONFIG).
  --timezone="UTC"      Timezone assigned to Diun ($TZ).
  --log-level="info"    Set log level ($LOG_LEVEL).
  --log-json            Enable JSON logging output ($LOG_JSON).
  --log-caller          Add file:line of the caller to log output ($LOG_CALLER).
  --test-notif          Test notification settings.
```

Following environment variables can be used in place of flags:

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `CONFIG`           |               | Diun configuration file |
| `TZ`               | `UTC`         | Timezone assigned |
| `LOG_LEVEL`        | `info`        | Log level output |
| `LOG_JSON`         | `false`       | Enable JSON logging output |
| `LOG_CALLER`       | `false`       | Enable to add `file:line` of the caller |

## Quick start with the Docker provider

Create a `docker-compose.yml` file that uses the official Diun image:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    volumes:
      - "./data:/data"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      - "DIUN_WATCH_WORKERS=20"
      - "DIUN_WATCH_SCHEDULE=*/30 * * * *"
      - "DIUN_PROVIDERS_DOCKER=true"
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
    restart: always
```

Here we use a minimal configuration to analyze **all running containers** (watch by default enabled) of your **local Docker** instance **every 30 minutes**.

That's it. Now you can launch Diun with the following command:

```shell
$ docker-compose up -d
```

If you prefer to rely on the configuration file instead of environment variables:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    volumes:
      - "./data:/data"
      - "./diun.yml:/diun.yml:ro"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "CONFIG=/diun.yml"
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
    restart: always
```

```yaml
# ./diun.yml

watch:
  workers: 20
  schedule: "*/30 * * * *"
  firstCheckNotif: false

providers:
  docker:
    watchByDefault: true
```
