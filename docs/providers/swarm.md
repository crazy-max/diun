# Swarm provider

## About

The Swarm provider allows you to analyze the services of your Swarm cluster to extract images found and check for
updates on the registry.

## Quick start

In this section we quickly go over a basic stack using your local swarm cluster.

Here we use our local Swarm provider with a minimum configuration to analyze labeled containers (watch by default
disabled).

Now let's create a simple stack for Diun:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    command: serve
    volumes:
      - "./data:/data"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      - "DIUN_WATCH_WORKERS=20"
      - "DIUN_WATCH_SCHEDULE=0 */6 * * *"
      - "DIUN_PROVIDERS_SWARM=true"
    deploy:
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.role == manager
```

And another one with a simple service:

```yaml
version: "3.5"

services:
  nginx:
    image: nginx
    ports:
      - target: 80
        published: 80
        protocol: udp
    deploy:
      mode: replicated
      replicas: 2
      labels:
        - "diun.enable=true"
        - "diun.watch_repo=true"
```

As an example we use [nginx](https://hub.docker.com/_/nginx/) Docker image. A few [labels](#docker-labels) are added
to configure the image analysis of this service for Diun. We can now start these 2 stacks:

```
docker stack deploy -c diun.yml diun
docker stack deploy -c nginx.yml nginx
```

Now take a look at the logs:

```
$ docker service logs -f diun_diun
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:19:57 CET INF Starting Diun dev
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:19:57 CET INF Starting Diun...
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:19:57 CET INF Found 1 image(s) to analyze provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:19:59 CET INF New image found id=myswarm image=docker.io/library/nginx:latest provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.4 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.8 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.7 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.9 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.6 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:01 CET INF New image found id=myswarm image=docker.io/library/nginx:1.9.5 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:mainline-alpine provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:alpine-perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:stable-perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:stable-alpine-perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:1 provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:mainline-alpine-perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:stable provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:mainline-perl provider=swarm
diun_diun.1.i1l4yuiafq6y@docker-desktop    | Sat, 14 Dec 2019 16:20:02 CET INF New image found id=myswarm image=docker.io/library/nginx:mainline provider=swarm
...
```

## Configuration

!!! hint
    Environment variable `DIUN_PROVIDERS_SWARM=true` can be used to enable this provider with default values.

### `endpoint`

Server address to connect to. Local if empty.

!!! example "File"
    ```yaml
    providers:
      swarm:
        endpoint: "unix:///var/run/docker.sock"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_SWARM_ENDPOINT`

#### `apiVersion`

Overrides the client version with the specified one.

!!! example "File"
    ```yaml
    providers:
      swarm:
        apiVersion: "1.39"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_SWARM_APIVERSION`

#### `tlsCertsPath`

Path to load the TLS certificates from.

!!! example "File"
    ```yaml
    providers:
      swarm:
        tlsCertsPath: "/certs/"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_SWARM_TLSCERTSPATH`

#### `tlsVerify`

Controls whether client verifies the server's certificate chain and hostname (default `true`).

!!! example "File"
    ```yaml
    providers:
      swarm:
        tlsVerify: true
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_SWARM_TLSVERIFY`

#### `watchByDefault`

Enable watch by default. If false, services that don't have `diun.enable=true` label will be ignored (default `false`).

!!! example "File"
    ```yaml
    providers:
      swarm:
        watchByDefault: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_SWARM_WATCHBYDEFAULT`


## Docker labels

You can configure more finely the way to analyze the image of your service through Docker labels:

| Name                          | Default       | Description   |
|-------------------------------|---------------|---------------|
| `diun.enable`                 |               | Set to true to enable image analysis of this service |
| `diun.regopt`                 |               | [Registry options](../config/regopts.md) name to use |
| `diun.watch_repo`             | `false`       | Watch all tags of this service image ([be careful](../faq.md#docker-hub-rate-limits) with this setting) |
| `diun.notify_on`              | `new;update`  | Semicolon separated list of status to be notified: `new`, `update`. |
| `diun.max_tags`               | `0`           | Maximum number of tags to watch if `diun.watch_repo` enabled. `0` means all of them |
| `diun.include_tags`           |               | Semicolon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo` |
| `diun.exclude_tags`           |               | Semicolon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo` |
| `diun.platform`               | _automatic_   | Platform to use (e.g. `linux/amd64`) |
