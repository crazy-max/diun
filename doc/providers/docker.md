# Docker provider

* [About](#about)
* [Quick start](#quick-start)
* [Provider configuration](#provider-configuration)
  * [Configuration file](#configuration-file)
  * [Environment variables](#environment-variables)
* [Docker labels](#docker-labels)

## About

The Docker provider allows you to analyze the containers of your Docker instance to extract images found and check for updates on the registry.

## Quick start

In this section we quickly go over a basic docker-compose file using your local docker provider.

Here we use a single Docker provider with a minimum configuration to analyze labeled containers (watch by default disabled), even stopped ones, of your local Docker instance.

Now let's create a simple docker-compose file with Diun and some simple services:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
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
      - "DIUN_PROVIDERS_DOCKER_WATCHSTOPPED=true"
    restart: always

  cloudflared:
    image: crazymax/cloudflared:latest
    ports:
      - target: 5053
        published: 5053
        protocol: udp
      - target: 49312
        published: 49312
        protocol: tcp
    environment:
      - "TZ=Europe/Paris"
      - "TUNNEL_DNS_UPSTREAM=https://1.1.1.1/dns-query,https://1.0.0.1/dns-query"
    labels:
      - "diun.enable=true"
      - "diun.watch_repo=true"
    restart: always
```

As an example we use [crazymax/cloudflared:latest](https://github.com/crazy-max/docker-cloudflared) Docker image. A few [labels](#docker-labels) are added to configure the image analysis of this container for Diun. Now start this composition with `docker-composes up -d` and take a look at the logs:

```
$ docker-compose logs -f
Attaching to bin_diun_1, cloudflared
cloudflared    | time="2019-12-14T15:30:07+01:00" level=info msg="Adding DNS upstream" url="https://1.1.1.1/dns-query"
cloudflared    | time="2019-12-14T15:30:07+01:00" level=info msg="Adding DNS upstream" url="https://1.0.0.1/dns-query"
cloudflared    | time="2019-12-14T15:30:07+01:00" level=info msg="Starting metrics server" addr="[::]:49312"
cloudflared    | time="2019-12-14T15:30:07+01:00" level=info msg="Starting DNS over HTTPS proxy server" addr="dns://0.0.0.0:5053"
diun_1         | Sat, 14 Dec 2019 15:30:07 CET INF Starting Diun v2.0.0
diun_1         | Sat, 14 Dec 2019 15:30:07 CET INF Found 1 image(s) to analyze provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:10 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:latest provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.9.0 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.9.1 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.9.2 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.10.1 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.10.4 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.10.2 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.11.0 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:12 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.11.3 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:13 CET INF New image found id=mydocker image=docker.io/crazymax/cloudflared:2019.11.2 provider=docker
diun_1         | Sat, 14 Dec 2019 15:30:13 CET INF Cron initialized with schedule */30 * * * *
diun_1         | Sat, 14 Dec 2019 15:30:13 CET INF Next run in 29 minutes (2019-12-14 16:00:00 +0100 CET)
```

## Provider configuration

### Configuration file

#### `endpoint`

Server address to connect to. Local if empty.

```yaml
providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
```

#### `apiVersion`

Overrides the client version with the specified one.

```yaml
providers:
  docker:
    apiVersion: "1.39"
```

#### `tlsCertsPath`

Path to load the TLS certificates from.

```yaml
providers:
  docker:
    tlsCertsPath: "/certs/"
```

#### `tlsVerify`

Controls whether client verifies the server's certificate chain and hostname (default `true`).

```yaml
providers:
  docker:
    tlsVerify: true
```

#### `watchByDefault`

Enable watch by default. If false, containers that don't have `diun.enable=true` label will be ignored (default `false`).

```yaml
providers:
  docker:
    watchByDefault: false
```

#### `watchStopped`

Include created and exited containers too (default `false`).

```yaml
providers:
  docker:
    watchStopped: false
```

### Environment variables

* `DIUN_PROVIDERS_DOCKER`
* `DIUN_PROVIDERS_DOCKER_ENDPOINT`
* `DIUN_PROVIDERS_DOCKER_APIVERSION`
* `DIUN_PROVIDERS_DOCKER_TLSCERTSPATH`
* `DIUN_PROVIDERS_DOCKER_TLSVERIFY`
* `DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT`
* `DIUN_PROVIDERS_DOCKER_WATCHSTOPPED`

## Docker labels

You can configure more finely the way to analyze the image of your container through Docker labels:

* `diun.enable`: Set to true to enable image analysis of this container.
* `diun.regopts_id`: Registry options ID from [`regopts`](../configuration.md#regopts) to use.
* `diun.watch_repo`: Watch all tags of this container image (default `false`).
* `diun.max_tags`: Maximum number of tags to watch if `diun.watch_repo` enabled. 0 means all of them (default `0`).
* `diun.include_tags`: Semi-colon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`.
* `diun.exclude_tags`: Semi-colon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`.
