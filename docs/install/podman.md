# Podman monitoring

## About

[Podman](https://podman.io) is now the default container runtime in some, *mainly RPM-based*, Linux distributions like RHEL.

Podman needs only a config tweak to work with diun as it is backwards-compatible with the Docker API

## Prerequisites

!!! warning
    Package `podman-docker` is not available in debian-based environnement, only native method is supported

At first you need to check if you have the `podman-docker` package installed and then follow the appropriate section :

You can use :

```shell
rpm -q podman-docker
```

### Native podman method

!!! warning
    Only install via `podman-compose`/`podman run` will work, due to file location change between podman and docker

!!! note
    You need [Podman Compose](https://github.com/containers/podman-compose) to process `docker-compose` files


Enable and start the socket :

```shell
# When monitoring root instance :
  sudo systemctl enable podman.socket
  sudo systemctl start podman.socket
# When monitoring user instance (connected as the user)
  systemctl enable --user podman.socket
  systemctl start --user podman.socket
```

Without `podman-docker`, you need to modfiy the socket location

Replace this line in `docker-compose.yml` or in `podman run` *(keep the `-v`)* :

```yaml
     - "/var/run/docker.sock:/var/run/docker.sock"
```

By *(without `-` on `podman run`)* :

```yaml
# When monitoring root instance :
     - "/run/podman/podman.sock:/var/run/docker.sock"
# When monitoring rootless podman :
     - "$XDG_RUNTIME_DIR/podman/podman.sock:/var/run/docker.sock"
     # You need to modify $XDG_RUNTIME_DIR by the real value of the variable
     # for the given user
```

`podman-compose` requires you to modify the container name to correctly start the pod

```diff
- container_name: diun
+ container_name: diun_app #or any name
```

Replacing the image name by the qualified name of the image is advised

```diff
- crazymax/diun:latest
+ docker.io/crazymax/diun:latest
```

#### Examples

##### Rootful (`root` user)


??? example "rootful docker-compose.yml"
    ```yaml
    version: "3.5"

    services:
      diun:
        image: docker.io/crazymax/diun:latest
        container_name: diun_app
        volumes:
          - "./data:/data"
          - "/run/podman/podman.sock:/var/run/docker.sock"
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

??? example "rootful podman run"
    ```shell
    podman run -d --name diun \
      -e "TZ=Europe/Paris" \
      -e "LOG_LEVEL=info" \
      -e "LOG_JSON=false" \
      -e "DIUN_WATCH_WORKERS=20" \
      -e "DIUN_WATCH_SCHEDULE=0 */6 * * *" \
      -e "DIUN_PROVIDERS_DOCKER=true" \
      -e "DIUN_PROVIDERS_DOCKER_WATCHSTOPPED=true" \
      -v "$(pwd)/data:/data" \
      -v "/run/podman/podman.sock:/var/run/docker.sock" \
      -l "diun.enable=true" \
      -l "diun.watch_repo=true" \
      docker.io/crazymax/diun:latest
    ```

##### Rootless

Grab the value of `$XDG_RUNTIME_DIR`

```shell
echo $XDG_RUNTIME_DIR
# Usually /run/user/1000, but can vary
```

With `/run/user/1000` as an example


??? example "rootless docker-compose.yml"
    ```yaml
    version: "3.5"

    services:
      diun:
        image: docker.io/crazymax/diun:latest
        container_name: diun_app
        volumes:
          - "./data:/data"
          - "/run/user/1000/podman/podman.sock:/var/run/docker.sock"
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

??? example "rootless podman run"
    ```shell
    podman run -d --name diun \
      -e "TZ=Europe/Paris" \
      -e "LOG_LEVEL=info" \
      -e "LOG_JSON=false" \
      -e "DIUN_WATCH_WORKERS=20" \
      -e "DIUN_WATCH_SCHEDULE=0 */6 * * *" \
      -e "DIUN_PROVIDERS_DOCKER=true" \
      -e "DIUN_PROVIDERS_DOCKER_WATCHSTOPPED=true" \
      -v "$(pwd)/data:/data" \
      -v "/run/user/1000/podman/podman.sock:/var/run/docker.sock" \
      -l "diun.enable=true" \
      -l "diun.watch_repo=true" \
      docker.io/crazymax/diun:latest
    ```

### Podman-docker method

!!! warning
    If you have this package installed, you will only be able to monitor `root` podman instance.  
    You will **NOT** be able to monitor any user *rootless* instance.  
    Therefore this is not the recommanded method.

Enable and start the socket with :

```shell
sudo systemctl enable podman.socket
sudo systemctl start podman.socket
```

And you're good to go with default configuration and binary install if needed
