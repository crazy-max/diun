# Configuration

* [Overview](#overview)
* [Configuration file](#configuration-file)
* [Reference](#reference)
  * [db](#db)
  * [watch](#watch)
  * [notif](#notif)
  * [regopts](#regopts)
  * [providers](#providers)

## Overview

There are two different ways to define static configuration options in Diun:

* In a [configuration file](#configuration-file)
* As environment variables

These ways are evaluated in the order listed above.

If no value was provided for a given option, a default value applies. Moreover, if an option has sub-options, and any of these sub-options is not specified, a default value will apply as well.

For example, the `DIUN_PROVIDERS_DOCKER` environment variable is enough by itself to enable the docker provider, even though sub-options like `DIUN_PROVIDERS_DOCKER_ENDPOINT` exist. Once positioned, this option sets (and resets) all the default values of the sub-options of `DIUN_PROVIDERS_DOCKER`.

## Configuration file

You can define a configuration file through the option `--config` with the following content:

```yaml
db:
  path: diun.db

watch:
  workers: 10
  schedule: "0 * * * *"
  firstCheckNotif: false

notif:
  amqp:
    host: localhost
    port: 5672
    username: guest
    password: guest
    queue: queue
  gotify:
    endpoint: http://gotify.foo.com
    token: Token123456
    priority: 1
    timeout: 10s
  mail:
    host: localhost
    port: 25
    ssl: false
    insecureSkipVerify: false
    from: diun@example.com
    to: webmaster@example.com
  rocketchat:
    endpoint: http://rocket.foo.com:3000
    channel: "#general"
    userID: abcdEFGH012345678
    token: Token123456
    timeout: 10s
  script:
      cmd: "myprogram"
      args:
        - "--anarg"
        - "another"
  slack:
    webhookURL: https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
  teams:
    webhookURL: https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
  telegram:
    token: aabbccdd:11223344
    chatIDs:
      - 123456789
      - 987654321
  webhook:
    endpoint: http://webhook.foo.com/sd54qad89azd5a
    method: GET
    headers:
      content-type: application/json
      authorization: Token123456
    timeout: 10s

regopts:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20s
  onemore:
    username: foo2
    password: bar2
    insecureTls: true

providers:
  docker:
    watchStopped: true
  swarm:
    watchByDefault: true
  file:
    directory: ./imagesdir
```

## Reference

### db

* `path`: Path to Bolt database file where images manifests are stored. (default `diun.db`)

You can also use the following environment variables:

* `DIUN_DB_PATH`

### watch

* `workers`: Maximum number of workers that will execute tasks concurrently. (default `10`)
* `schedule`: [CRON expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) to schedule Diun watcher. (default `0 * * * *`)
* `firstCheckNotif`: Send notification at the very first analysis of an image. (default `false`)

You can also use the following environment variables:

* `DIUN_WATCH_WORKERS`
* `DIUN_WATCH_SCHEDULE`
* `DIUN_WATCH_FIRSTCHECKNOTIF`

### notif

* [amqp](notifications.md#amqp)
* [gotify](notifications.md#gotify)
* [mail](notifications.md#mail)
* [rocketchat](notifications.md#rocketchat)
* [script](notifications.md#script)
* [slack](notifications.md#slack--mattermost)
* [teams](notifications.md#teams)
* [telegram](notifications.md#telegram)
* [webhook](notifications.md#webhook)

### regopts

* `username`: Registry username.
* `usernameFile`: Use content of secret file as registry username if `username` not defined.
* `password`: Registry password.
* `passwordFile`: Use content of secret file as registry password if `password` not defined.
* `timeout`: Timeout is the maximum amount of time for the TCP connection to establish. (default `10s`)
* `insecureTls`: Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification. (default `false`)

You can also use the following environment variables:

* `DIUN_REGOPTS_<NAME>_USERNAME`
* `DIUN_REGOPTS_<NAME>_USERNAMEFILE`
* `DIUN_REGOPTS_<NAME>_PASSWORD`
* `DIUN_REGOPTS_<NAME>_PASSWORDFILE`
* `DIUN_REGOPTS_<NAME>_TIMEOUT`
* `DIUN_REGOPTS_<NAME>_INSECURETLS`

### providers

* [docker](providers/docker.md)
* [swarm](providers/swarm.md)
* [file](providers/file.md)
