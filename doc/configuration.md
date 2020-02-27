# Configuration

* [Overview](#overview)
* [Reference](#reference)
  * [db](#db)
  * [watch](#watch)
  * [notif](#notif)
  * [regopts](#regopts)
  * [providers](#providers)

## Overview

Here is a YAML structure example:

```yml
db:
  path: diun.db

watch:
  workers: 10
  schedule: "0 * * * *"
  first_check_notif: false

notif:
  gotify:
    enable: false
    endpoint: http://gotify.foo.com
    token: Token123456
    priority: 1
    timeout: 10
  mail:
    enable: false
    host: localhost
    port: 25
    ssl: false
    insecure_skip_verify: false
    username:
    password:
    from:
    to:
  slack:
    enable: false
    webhook_url: https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
  telegram:
    enable: false
    token: aabbccdd:11223344
    chat_ids:
      - 123456789
      - 987654321
  webhook:
    enable: false
    endpoint: http://webhook.foo.com/sd54qad89azd5a
    method: GET
    headers:
      Content-Type: application/json
      Authorization: Token123456
    timeout: 10

regopts:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20
  onemore:
    username: foo2
    password: bar2
    insecure_tls: true

providers:
  docker:
    # Watch only labeled containers on local Docker engine
    local:
      watch_stopped: true
    # Watch all containers on 10.0.0.1:2375
    remote:
      endpoint: tcp://10.0.0.1:2375
      watch_by_default: true
  swarm:
    # Watch all services on local Swarm cluster
    myswarm:
      watch_by_default: true
  static:
    # Watch latest tag of crazymax/nextcloud image on docker.io (DockerHub) with registry ID 'someregistryoptions'.
    - name: docker.io/crazymax/nextcloud:latest
      regopts_id: someregistryoptions
    # Watch 4.0.0 tag of jfrog/artifactory-oss image on frog-docker-reg2.bintray.io (Bintray) with registry ID 'onemore'.
    - name: jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0
      regopts_id: onemore
    # Watch coreos/hyperkube image on quay.io (Quay) and assume latest tag.
    - name: quay.io/coreos/hyperkube
    # Watch crazymax/swarm-cronjob image and assume docker.io registry and latest tag.
    # Only include tags matching regexp ^1\.2\..*
    - name: crazymax/swarm-cronjob
      watch_repo: true
      include_tags:
        - ^1\.2\..*
    # Watch portainer/portainer image on docker.io (DockerHub) and assume latest tag
    # Only watch latest 10 tags and include tags matching regexp ^(0|[1-9]\d*)\..*
    - name: docker.io/portainer/portainer
      watch_repo: true
      max_tags: 10
      include_tags:
        - ^(0|[1-9]\d*)\..*
    # Watch alpine image (library) and assume docker.io registry and latest tag.
    # Only check linux/arm64v8 image
    - name: alpine
      watch_repo: true
      os: linux
      arch: arm64v8
```

## Reference

### db

* `path`: Path to Bolt database file where images manifests are stored (default: `diun.db`). Environment var `DIUN_DB` override this value.

### watch

* `workers`: Maximum number of workers that will execute tasks concurrently. _Optional_. (default: `10`).
* `schedule`: [CRON expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) to schedule Diun watcher. _Optional_. (default: `0 * * * *`).
* `first_check_notif`: Send notification at the very first analysis of an image. _Optional_. (default: `false`).

### notif

* `gotify`
  * `enable`: Enable gotify notification (default: `false`).
  * `endpoint`: Gotify base URL (e.g. `http://gotify.foo.com`). **required**
  * `token`: Application token. **required**
  * `priority`: The priority of the message.
  * `timeout`: Timeout specifies a time limit for the request to be made. (default: `10`).

* `mail`
  * `enable`: Enable email reports (default: `false`).
  * `host`: SMTP server host (default: `localhost`). **required**
  * `port`: SMTP server port (default: `25`). **required**
  * `ssl`: SSL defines whether an SSL connection is used. Should be false in most cases since the auth mechanism should use STARTTLS (default: `false`).
  * `insecure_skip_verify`: Controls whether a client verifies the server's certificate chain and hostname (default: `false`).
  * `username`: SMTP username.
  * `username_file`: Use content of secret file as SMTP username if `username` not defined.
  * `password`: SMTP password.
  * `password_file`: Use content of secret file as SMTP password if `password` not defined.
  * `from`: Sender email address. **required**
  * `to`: Recipient email address. **required**

* `slack`
  * `enable`: Enable slack notification (default: `false`).
  * `webhook_url`: Slack [incoming webhook URL](https://api.slack.com/messaging/webhooks). **required**

* `telegram`
  * `enable`: Enable Telegram notification (default: `false`).
  * `token`: Telegram bot token. **required**
  * `chat_ids`: List of chat IDs to send notifications to. **required**

* `webhook`
  * `enable`: Enable webhook notification (default: `false`).
  * `endpoint`: URL of the HTTP request. **required**
  * `method`: HTTP method (default: `GET`). **required**
  * `headers`: Map of additional headers to be sent.
  * `timeout`: Timeout specifies a time limit for the request to be made. (default: `10`).

### regopts

* `username`: Registry username.
* `username_file`: Use content of secret file as registry username if `username` not defined.
* `password`: Registry password.
* `password_file`: Use content of secret file as registry password if `password` not defined.
* `timeout`: Timeout is the maximum amount of time for the TCP connection to establish. 0 means no timeout (default: `10`).
* `insecure_tls`: Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification (default: `false`).

### providers

* `docker`: Map of Docker standalone engines to watch
  * `<key>`: An unique identifier for this provider.
    * `endpoint`: Server address to connect to. Local if empty. _Optional_
    * `api_version`: Overrides the client version with the specified one. _Optional_
    * `tls_certs_path`: Path to load the TLS certificates from. _Optional_
    * `tls_verify`: Controls whether client verifies the server's certificate chain and hostname (default: `true`).
    * `watch_by_default`: Enable watch by default. If false, containers that don't have `diun.enable=true` label will be ignored (default: `false`).
    * `watch_stopped`: Include created and exited containers too (default: `false`).

* `swarm`: Map of Docker Swarm to watch
  * `<key>`: An unique identifier for this provider.
    * `endpoint`: Server address to connect to. Local if empty. _Optional_
    * `api_version`: Overrides the client version with the specified one. _Optional_
    * `tls_certs_path`: Path to load the TLS certificates from. _Optional_
    * `tls_verify`: Controls whether client verifies the server's certificate chain and hostname (default: `true`).
    * `watch_by_default`: Enable watch by default. If false, services that don't have `diun.enable=true` label will be ignored (default: `false`).

* `static`: Slice of static image to watch
  * `name`: Docker image name to watch using `registry/path:tag` format. If registry is omitted, `docker.io` will be used and if tag is omitted, `latest` will be used. **required**
  * `os`: OS to use. _Optional_. (default: `linux`).
  * `arch`: Architecture to use. _Optional_. (default: `amd64`).
  * `regopts_id`: Registry options ID from `regopts` to use.
  * `watch_repo`: Watch all tags of this `image` repository (default: `false`).
  * `max_tags`: Maximum number of tags to watch if `watch_repo` enabled. 0 means all of them (default: `0`).
  * `include_tags`: List of regular expressions to include tags. Can be useful if you enable `watch_repo`.
  * `exclude_tags`: List of regular expressions to exclude tags. Can be useful if you enable `watch_repo`.
