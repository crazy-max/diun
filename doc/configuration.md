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
  amqp:
    enable: false
    host: localhost
    port: 5672
    username: guest
    password: guest
    exchange: 
    queue: queue

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
  file:
    # Watch images from filename ./myimages.yml
    filename: ./myimages.yml
    # Watch images from directory ./imagesdir
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

* `rocketchat`
  * `enable`: Enable Rocket.Chat notification (default: `false`).
  * `endpoint`: Rocket.Chat base URL (e.g. `http://rocket.foo.com:3000`). **required**
  * `channel`: Channel name with the prefix in front of it. **required**
  * `user_id`: User ID. **required**
  * `token`: Authentication token. **required**
  * `timeout`: Timeout specifies a time limit for the request to be made. (default: `10`).

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

* `amqp`
  * `enable`: Enable AMQP notifications (default: `false`).
  * `host`: AMQP server host (default: `localhost`). **required**
  * `port`: AMQP server port (default: `5672`). **required**
  * `username`: AMQP username. **required**  
  * `password`: AMQP password. **required**
  * `exchange`: Name of the exchange the message will be sent to.
  * `queue`: Name of the queue the message will be sent to. **required**
  
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
