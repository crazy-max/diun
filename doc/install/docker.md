# Installation with Docker

Diun provides automatically updated Docker :whale: images within [Docker Hub](https://hub.docker.com/r/crazymax/diun) and [Quay](https://quay.io/repository/crazymax/diun). It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

Environment variables can be used within your container :

* `TZ` : Timezone assigned
* `LOG_LEVEL` : Log level output (default `info`)
* `LOG_JSON`: Enable JSON logging output (default `false`)
* `LOG_CALLER`: Enable to add file:line of the caller (default `false`)

Docker compose is the recommended way to run this image. Copy the content of folder [.res/compose](../../.res/compose) in `/opt/diun/` on your host for example. Edit the compose and config file with your preferences and run the following commands :

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
