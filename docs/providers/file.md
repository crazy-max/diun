# File provider

## About

The file provider lets you define Docker images to analyze through a YAML file or a directory.

## Example

Register the file provider:

```yaml
db:
  path: diun.db

watch:
  workers: 20
  schedule: "0 */6 * * *"

regopts:
  - name: "myregistry"
    username: fii
    password: bor
    timeout: 5s
  - name: "docker.io/crazymax"
    selector: image
    username: fii
    password: bor
  - name: "docker.io"
    selector: image
    username: foo
    password: bar

providers:
  file:
    filename: /path/to/config.yml
```

```yaml
### /path/to/config.yml

# Watch latest tag of crazymax/nextcloud image on docker.io (DockerHub)
# with registry options named 'docker.io/crazymax' (image selector).
- name: docker.io/crazymax/nextcloud:latest

# Watch 4.0.0 tag of jfrog/artifactory-oss image on frog-docker-reg2.bintray.io (Bintray)
# with registry options named 'myregistry' (name selector).
- name: jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0
  regopt: myregistry

# Watch coreos/hyperkube image on quay.io (Quay) and assume latest tag.
- name: quay.io/coreos/hyperkube

# Watch crazymax/swarm-cronjob image and assume docker.io registry and latest tag
# with registry options named 'docker.io/crazymax' (image selector).
# Only include tags matching regexp ^1\.2\..*
- name: crazymax/swarm-cronjob
  watch_repo: true
  include_tags:
    - ^1\.2\..*

# Watch portainer/portainer image on docker.io (DockerHub) and assume latest tag
# with registry options named 'docker.io' (image selector).
# Only watch latest 10 tags and include tags matching regexp ^\d+\.\d+\..*
- name: docker.io/portainer/portainer
  watch_repo: true
  max_tags: 10
  include_tags:
    - ^\d+\.\d+\..*

# Watch alpine image (library) and assume docker.io registry and latest tag
# with registry options named 'docker.io' (image selector).
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
  schedule: "0 */6 * * *"

regopts:
  - name: "docker.bintray.io"
    selector: image
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
```

Here we want to analyze all tags of `crazymax/cloudflared` and `docker.bintray.io/jfrog/xray-mongo:3.2.6` tag.
Now let's start Diun:

```
$ diun serve --config diun.yml
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
Sat, 14 Dec 2019 15:32:28 UTC INF Cron initialized with schedule 0 */6 * * *
Sat, 14 Dec 2019 15:32:28 UTC INF Next run in 31 seconds (2019-12-14 15:33:00 +0000 UTC)
```

## Configuration

### `filename`

Defines the path to the [configuration file](#yaml-configuration-file).

!!! warning
    `filename` and `directory` are mutually exclusive

!!! example "File"
    ```yaml
    providers:
      file:
        filename: /path/to/config/conf.yml
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_FILE_FILENAME`

### `directory`

Defines the path to the directory that contains the [configuration files](#yaml-configuration-file) (`*.yml` or `*.yaml`).

!!! warning
    `filename` and `directory` are mutually exclusive

!!! example "File"
    ```yaml
    providers:
      file:
        directory: /path/to/config
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_FILE_DIRECTORY`

## YAML configuration file

The configuration file(s) defines a slice of images to analyze with the following fields:

| Name                          | Default                          | Description   |
|-------------------------------|----------------------------------|---------------|
| `name`                        | `latest`                         | Docker image name to watch using `registry/path:tag` format. If registry omitted, `docker.io` will be used and if tag omitted, `latest` will be used |
| `regopt`                      |                                  | [Registry options](../config/regopts.md) name to use |
| `watch_repo`                  | `false`                          | Watch all tags of this image ([be careful](../faq.md#docker-hub-rate-limits) with this setting) |
| `max_tags`                    | `0`                              | Maximum number of tags to watch if `watch_repo` enabled. `0` means all of them |
| `include_tags`                |                                  | List of regular expressions to include tags. Can be useful if you enable `watch_repo` |
| `exclude_tags`                |                                  | List of regular expressions to exclude tags. Can be useful if you enable `watch_repo` |
| `platform.os`                 | _automatic_                      | Operating system to use as custom platform |
| `platform.arch`               | _automatic_                      | CPU architecture to use as custom platform |
| `platform.variant`            | _automatic_                      | Variant of the CPU to use as custom platform |
