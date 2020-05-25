# Swarm provider

* [About](#about)
* [Quick start](#quick-start)
* [Provider configuration](#provider-configuration)
* [Docker labels](#docker-labels)

## About

The Swarm provider allows you to analyze the services of your Swarm cluster to extract images found and check for updates on the registry.

## Quick start

In this section we quickly go over a basic stack using your local swarm cluster.

First of all, let's create a Diun configuration we named `diun.yml`:

```yml
watch:
  workers: 20
  schedule: "*/30 * * * *"

providers:
  swarm:
```

Here we use our local Swarm provider with a minimum configuration to analyze labeled containers (watch by default disabled).

Now let's create a simple stack for Diun:

```yml
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
    deploy:
      placement:
        constraints:
          - node.role == manager
```

And another one with a simple service:

```yml
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

As an example we use [nginx](https://hub.docker.com/_/nginx/) Docker image. A few [labels](#docker-labels) are added to configure the image analysis of this service for Diun. We can now start these 2 stacks:

```
docker stack deploy -c diun.yml diun
docker stack deploy -c nginx.yml nginx
```

And watch logs of Diun service:

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

## Provider configuration

* `endpoint`: Server address to connect to. Local if empty.
* `api_version`: Overrides the client version with the specified one.
* `tls_certs_path`: Path to load the TLS certificates from.
* `tls_verify`: Controls whether client verifies the server's certificate chain and hostname (default: `true`).
* `watch_by_default`: Enable watch by default. If false, services that don't have `diun.enable=true` label will be ignored (default: `false`).

## Docker labels

You can configure more finely the way to analyze the image of your service through Docker labels:

* `diun.enable`: Set to true to enable image analysis of this container. Required if `watch_by_default` is disabled for this provider.
* `diun.regopts_id`: Registry options ID from [`regopts`](../configuration.md#regopts) to use.
* `diun.watch_repo`: Watch all tags of this container image (default: `false`).
* `diun.max_tags`: Maximum number of tags to watch if `diun.watch_repo` enabled. 0 means all of them (default: `0`).
* `diun.include_tags`: Semi-colon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`.
* `diun.exclude_tags`: Semi-colon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`.
