# Getting started

* [Diun CLI](#diun-cli)
* [Run with the Docker provider](#run-with-the-docker-provider)

## Diun CLI

```
$ ./diun --help
Usage: diun

Docker image update notifier. More info: https://github.com/crazy-max/diun

Flags:
  --help                Show context-sensitive help.
  --version
  --config=STRING       Diun configuration file ($CONFIG).
  --timezone="UTC"      Timezone assigned to Diun ($TZ).
  --log-level="info"    Set log level ($LOG_LEVEL).
  --log-json            Enable JSON logging output ($LOG_JSON).
  --log-caller          Add file:line of the caller to log output ($LOG_CALLER).
  --test-notif          Test notification settings.
```

Following environment variables can be used in place of flags:

* `CONFIG`: Diun configuration file
* `TZ`: Timezone assigned (default `UTC`)
* `LOG_LEVEL`: Log level output (default `info`)
* `LOG_JSON`: Enable JSON logging output (default `false`)
* `LOG_CALLER`: Enable to add `file:line` of the caller (default `false`)

## Run with the Docker provider

Create a `docker-compose.yml` file that uses the official Diun image:

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
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
    restart: always
```

Here we use a minimal configuration to analyze **all running containers** (watch by default enabled) of your **local Docker** instance **every 30 minutes**.

That's it. Now you can launch Diun with the following command:

```shell
$ docker-compose up -d
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
      - "CONFIG=/diun.yml"
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
    restart: always
```

```yaml
# ./diun.yml

watch:
  workers: 20
  schedule: "*/30 * * * *"
  firstCheckNotif: false

providers:
  docker:
    watchByDefault: true
```
