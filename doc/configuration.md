# Configuration

* [Overview](#overview)
* [Reference](#reference)
  * [db](#db)
  * [watch](#watch)
  * [notif](#notif)
  * [regopts](#regopts)
  * [providers](#providers)

## Overview

```yml
db:
  path: diun.db

watch:
  workers: 10
  schedule: "0 * * * *"
  first_check_notif: false

notif:
  amqp:
    enable: false
    host: localhost
    port: 5672
    username: guest
    password: guest
    exchange: 
    queue: queue
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
  rocketchat:
    enable: false
    endpoint: http://rocket.foo.com:3000
    channel: "#general"
    user_id: abcdEFGH012345678
    token: Token123456
    timeout: 10
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
    watch_stopped: true
  swarm:
    watch_by_default: true
  file:
    directory: ./imagesdir
```

## Reference

### db

* `path`: Path to Bolt database file where images manifests are stored (default: `diun.db`). Environment var `DIUN_DB` override this value.

### watch

* `workers`: Maximum number of workers that will execute tasks concurrently (default: `10`).
* `schedule`: [CRON expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) to schedule Diun watcher (default: `0 * * * *`).
* `first_check_notif`: Send notification at the very first analysis of an image. (default: `false`).

### notif

* [amqp](notifications.md#amqp)
* [gotify](notifications.md#gotify)
* [mail](notifications.md#mail)
* [rocketchat](notifications.md#rocketchat)
* [slack](notifications.md#slack)
* [telegram](notifications.md#telegram)
* [webhook](notifications.md#webhook)

### regopts

* `username`: Registry username.
* `username_file`: Use content of secret file as registry username if `username` not defined.
* `password`: Registry password.
* `password_file`: Use content of secret file as registry password if `password` not defined.
* `timeout`: Timeout is the maximum amount of time for the TCP connection to establish. 0 means no timeout (default: `10`).
* `insecure_tls`: Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification (default: `false`).

### providers

* [docker](providers/docker.md)
* [swarm](providers/swarm.md)
* [file](providers/file.md)
