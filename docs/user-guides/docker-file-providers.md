# Docker + File providers

It is possible to use several providers at the same time with Diun. This can be particularly useful if you want to
analyze some images that you don't manage through a container.

In this section we quickly go over a basic docker-compose file to run Diun using the [docker](../providers/docker.md)
and [file](../providers/file.md) providers.

## Setup

Create a `docker-compose.yml` file that uses the official Diun image:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    volumes:
      - "./data:/data"
      - "./custom-images.yml:/custom-images.yml:ro"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      - "DIUN_WATCH_WORKERS=20"
      - "DIUN_WATCH_SCHEDULE=0 */6 * * *"
      - "DIUN_PROVIDERS_DOCKER=true"
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
      - "DIUN_PROVIDERS_FILE_FILENAME=/custom-images.yml"
    restart: always
```

```yaml
# /custom-images.yml
- name: ghcr.io/crazy-max/diun
  watch_repo: true
- name: alpine
  watch_repo: true
- name: debian:stretch-slim
- name: nginx:stable-alpine
- name: traefik
  watch_repo: true
  include_tags:
    - ^(0|[1-9]\d*)\..*-alpine
```

Here we use a minimal configuration to analyze **all running containers** (watch by default enabled) of
your **local Docker** instance with the [Docker provider](../providers/docker.md) and also **custom images**
through the [File provider](../providers/file.md) **every 6 hours**.

That's it. Now you can launch Diun with the following command:

```shell
$ docker-compose up -d
```
