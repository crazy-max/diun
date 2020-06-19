# Installation with Docker

Diun provides automatically updated Docker :whale: images within [Docker Hub](https://hub.docker.com/r/crazymax/diun).
It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

Following platforms for this image are available:

```shell
$ docker run --rm mplatform/mquery crazymax/diun:latest
Image: crazymax/diun:latest
 * Manifest List: Yes
 * Supported platforms:
   - linux/amd64
   - linux/arm/v6
   - linux/arm/v7
   - linux/arm64
   - linux/386
   - linux/ppc64le
   - linux/s390x
```

## Volumes

| Path               | Description   |
|--------------------|---------------|
| `/data`            | Contains bbolt database which retains Docker images manifests |

## Usage

Docker compose is the recommended way to run this image. Copy the following `docker-compose.yml` in `/opt/diun/` on your host for example:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    container_name: diun
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
    labels:
      - "diun.enable=true"
      - "diun.watch_repo=true"
    restart: always
```

Edit this example with your preferences and run the following commands to bring up Diun:

```shell
$ docker-compose up -d
$ docker-compose logs -f
```

Or use the following command:

```shell
$ docker run -d --name diun \
  -e "TZ=Europe/Paris" \
  -e "LOG_LEVEL=info" \
  -e "LOG_JSON=false" \
  -e "DIUN_WATCH_WORKERS=20" \
  -e "DIUN_WATCH_SCHEDULE=*/30 * * * *" \
  -e "DIUN_PROVIDERS_DOCKER=true" \
  -e "DIUN_PROVIDERS_DOCKER_WATCHSTOPPED=true" \
  -v "$(pwd)/data:/data" \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  -l "diun.enable=true" \
  -l "diun.watch_repo=true" \
  crazymax/diun:latest
```

To upgrade your installation to the latest release:

```shell
$ docker-compose pull
$ docker-compose up -d
```
