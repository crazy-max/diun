# File provider

* [About](#about)
* [Example](#example)
* [Quick start](#quick-start)
* [Provider configuration](#provider-configuration)
  * [Configuration file](#configuration-file)
  * [Environment variables](#environment-variables)
* [YAML configuration file](#yaml-configuration-file)

## About

The file provider lets you define Docker images to analyze through a YAML file or a directory.

## Example

Register the file provider:

```yaml
db:
  path: diun.db

watch:
  workers: 20
  schedule: "* * * * *"

regopts:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20
  onemore:
    username: foo2
    password: bar2
    insecureTls: true

providers:
  file:
    filename: /path/to/config.yml
```

```yaml
### /path/to/config.yml

# Watch latest tag of crazymax/nextcloud image on docker.io (DockerHub) with registry ID 'someregistryoptions'.
- name: docker.io/crazymax/nextcloud:latest
  regopts_id: someregistryoptions

# Watch 4.0.0 tag of jfrog/artifactory-oss image on frog-docker-reg2.bintray.io (Bintray) with registry ID 'onemore'.
- name: jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0
  regopts_id: onemore

# Watch coreos/hyperkube image on quay.io (Quay) and assume latest tag.
- name: quay.io/coreos/hyperkube

# Watch crazymax/swarm-cronjob image and assume docker.io registry and latest tag.
# Only include tags matching regexp ^1\.2\..*
- name: crazymax/swarm-cronjob
  watch_repo: true
  include_tags:
    - ^1\.2\..*

# Watch portainer/portainer image on docker.io (DockerHub) and assume latest tag
# Only watch latest 10 tags and include tags matching regexp ^(0|[1-9]\d*)\..*
- name: docker.io/portainer/portainer
  watch_repo: true
  max_tags: 10
  include_tags:
    - ^(0|[1-9]\d*)\..*

# Watch alpine image (library) and assume docker.io registry and latest tag.
# Force linux/arm64/v8 platform for this image
- name: alpine
  watch_repo: true
  platform:
    os: linux
    arch: arm64
    variant: v8
```

## Quick start

Let's take a look with a simple example:

```yaml
db:
  path: diun.db

watch:
  workers: 20
  schedule: "* * * * *"

regopts:
  jfrog:
    username: foo
    password: bar

providers:
  file:
    filename: /path/to/config.yml
```

```yaml
# /path/to/config.yml
- name: crazymax/cloudflared
  watch_repo: true
- name: docker.bintray.io/jfrog/xray-mongo:3.2.6
  regopts_id: jfrog
```

Here we want to analyze all tags of `crazymax/cloudflared` and `docker.bintray.io/jfrog/xray-mongo:3.2.6` tag. Now let's start Diun:

```
$ diun --config diun.yml
Sat, 14 Dec 2019 15:32:23 UTC INF Starting Diun 2.0.0
Sat, 14 Dec 2019 15:32:23 UTC INF Found 2 image(s) to analyze... provider=file
Sat, 14 Dec 2019 15:32:25 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:latest provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.3 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.0 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.1 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.0 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.2 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.2 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.2 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.1 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.4 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=docker.bintray.io/jfrog/xray-mongo:3.2.6 image=docker.bintray.io/jfrog/xray-mongo:3.2.6 provider=file
Sat, 14 Dec 2019 15:32:28 UTC INF Cron initialized with schedule * * * * *
Sat, 14 Dec 2019 15:32:28 UTC INF Next run in 31 seconds (2019-12-14 15:33:00 +0000 UTC)
```

## Provider configuration

### Configuration file

#### filename

Defines the path to the [configuration file](#yaml-configuration-file).

> :warning: `filename` and `directory` are mutually exclusive.

```yaml
providers:
  file:
    filename: /path/to/config/conf.yml
```

#### directory

Defines the path to the directory that contains the [configuration files](#yaml-configuration-file) (`*.yml` or `*.yaml`).

> :warning: `filename` and `directory` are mutually exclusive.

```yaml
providers:
  file:
    directory: /path/to/config
```

### Environment variables

* `DIUN_PROVIDERS_FILE_DIRECTORY`
* `DIUN_PROVIDERS_FILE_FILENAME`

## YAML configuration file

The configuration file(s) defines a slice of images to analyze with the following fields:

* `name`: Docker image name to watch using `registry/path:tag` format. If registry omitted, `docker.io` will be used and if tag omitted, `latest` will be used. **required**
* `regopts_id`: Registry options ID from [`regopts`](../configuration.md#regopts) to use.
* `watch_repo`: Watch all tags of this `image` repository (default `false`).
* `max_tags`: Maximum number of tags to watch if `watch_repo` enabled. 0 means all of them (default `0`).
* `include_tags`: List of regular expressions to include tags. Can be useful if you enable `watch_repo`.
* `exclude_tags`: List of regular expressions to exclude tags. Can be useful if you enable `watch_repo`.
* `platform`: Check a custom platform. (default will retrieve platform dynamically based on your operating system).
  * `os`: Operating system to use.
  * `arch`: CPU architecture to use.
  * `variant`: Variant of the CPU to use.
