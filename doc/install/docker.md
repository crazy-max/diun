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

## Environment variables

* `TZ` : Timezone assigned
* `LOG_LEVEL` : Log level output (default `info`)
* `LOG_JSON`: Enable JSON logging output (default `false`)
* `LOG_CALLER`: Enable to add file:line of the caller (default `false`)

## Volumes

* `/data` : Contains bbolt database which retains Docker images manifests

> :warning: Note that the volume should be owned by uid `1000` and gid `1000`. If you don't give the volume correct permissions, the container may not start.

## Usage

Docker compose is the recommended way to run this image. Copy the content of folder [.res/compose](../../.res/compose) in `/opt/diun/` on your host for example. Edit the compose and config file with your preferences and run the following commands:

```
docker-compose up -d
docker-compose logs -f
```

Or use the following command :

```
$ docker run -d --name diun \
  -e "TZ=Europe/Paris" \
  -e "LOG_LEVEL=info" \
  -e "LOG_JSON=false" \
  -v "$(pwd)/data:/data" \
  -v "$(pwd)/diun.yml:/diun.yml:ro" \
  crazymax/diun:latest
```

To upgrade your installation to the latest release:

```
docker-compose pull
docker-compose up -d
```
