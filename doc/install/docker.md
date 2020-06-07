# Installation with Docker

Diun provides automatically updated Docker :whale: images within [Docker Hub](https://hub.docker.com/r/crazymax/diun). It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

Following platforms for this image are available:

```
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

* `/data`: Contains bbolt database which retains Docker images manifests

## Usage

Docker compose is the recommended way to run this image. Copy the content of folder [.res/compose](../../.res/compose) in `/opt/diun/` on your host for example. Edit the compose file with your preferences and run the following commands:

```
docker-compose up -d
docker-compose logs -f
```

Or use the following command:

```
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
  crazymax/diun:latest
```

To upgrade your installation to the latest release:

```
docker-compose pull
docker-compose up -d
```
