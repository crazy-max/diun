# Installation from Docker image

## About

Diun provides automatically updated Docker :whale: images within several registries:

| Registry                                                                                         | Image                           |
|--------------------------------------------------------------------------------------------------|---------------------------------|
| [Docker Hub](https://hub.docker.com/r/crazymax/diun/)                             | `crazymax/diun`                 |
| [GitHub Container Registry](https://github.com/users/crazy-max/packages/container/package/diun)  | `ghcr.io/crazy-max/diun`        |

It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

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
```

## Volumes

| Path               | Description   |
|--------------------|---------------|
| `/data`            | Contains bbolt database which retains Docker images manifests |

## Usage

!!! note
    This reference setup guides users through the setup based on `docker-compose` and the
    [Docker provider](../providers/docker.md), but the installation of `docker-compose` is out of scope of this
    documentation. To install `docker-compose` itself, follow the official
    [install instructions](https://docs.docker.com/compose/install/).
    
    You can also use the [Swarm](../providers/swarm.md) or [Kubernetes](../providers/kubernetes.md) providers
    if you don't want to use `docker-compose`.

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
      - "DIUN_WATCH_SCHEDULE=0 */6 * * *"
      - "DIUN_PROVIDERS_DOCKER=true"
      - "DIUN_PROVIDERS_DOCKER_WATCHSTOPPED=true"
    labels:
      - "diun.enable=true"
      - "diun.watch_repo=true"
    restart: always
```

!!! note
    If you are running [podman](https://podman.io/), please check the appropriate [page](./podman.md)

Edit this example with your preferences and run the following commands to bring up Diun:

```shell
docker-compose up -d
docker-compose logs -f
```

Or use the following command:

```shell
docker run -d --name diun \
  -e "TZ=Europe/Paris" \
  -e "LOG_LEVEL=info" \
  -e "LOG_JSON=false" \
  -e "DIUN_WATCH_WORKERS=20" \
  -e "DIUN_WATCH_SCHEDULE=0 */6 * * *" \
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
docker-compose pull
docker-compose up -d
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
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
    restart: always
```

```yaml
# ./diun.yml

watch:
  workers: 20
  schedule: "0 */6 * * *"

providers:
  docker:
    watchStopped: true
```
