# Configuration

## Overview

There are two different ways to define configuration in Diun:

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
    insecureTLS: true

providers:
  docker:
    watchStopped: true
  swarm:
    watchByDefault: true
  kubernetes:
    namespaces:
      - default
      - production
  file:
    directory: ./imagesdir
```

## Reference

* [db](db.md)
* [watch](watch.md)
* notif
    * [amqp](../notif/amqp.md)
    * [gotify](../notif/amqp.md)
    * [mail](../notif/amqp.md)
    * [rocketchat](../notif/amqp.md)
    * [script](../notif/amqp.md)
    * [slack](../notif/amqp.md)
    * [teams](../notif/amqp.md)
    * [telegram](../notif/amqp.md)
    * [webhook](../notif/amqp.md)
* [regopts](regopts.md)
* providers
    * [docker](../providers/docker.md)
    * [file](../providers/file.md)
    * [kubernetes](../providers/kubernetes.md)
    * [swarm](../providers/swarm.md)
