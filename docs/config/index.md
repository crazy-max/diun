# Configuration

## Overview

There are two different ways to define configuration in Diun:

* In a [configuration file](#configuration-file)
* As [environment variables](#environment-variables)

These ways are evaluated in the order listed above.

If no value was provided for a given option, a default value applies. Moreover, if an option has sub-options, and any of these sub-options is not specified, a default value will apply as well.

For example, the `DIUN_PROVIDERS_DOCKER` environment variable is enough by itself to enable the docker provider, even though sub-options like `DIUN_PROVIDERS_DOCKER_ENDPOINT` exist. Once positioned, this option sets (and resets) all the default values of the sub-options of `DIUN_PROVIDERS_DOCKER`.

## Configuration file

At startup, Diun searches for a file named `diun.yml` (or `diun.yaml`) in:

* `/etc/diun/`
* `$XDG_CONFIG_HOME/`
* `$HOME/.config/`
* `.` _(the working directory)_

You can override this using the [`--config` flag or `CONFIG` env var with `serve` command](../usage/command-line.md#serve).

??? example "diun.yml"
    ```yaml
    db:
      path: diun.db

    watch:
      workers: 10
      schedule: "0 */6 * * *"
      jitter: 30s
      firstCheckNotif: false
      runOnStartup: true

    defaults:
      watchRepo: false
      notifyOn:
        - new
        - update
      sortTags: reverse

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
        to:
          - webmaster@example.com
          - me@example.com
      ntfy:
        endpoint: https://ntfy.sh
        topic: diun-acce65a0-b777-46f9-9a11-58c67d1579c4
        priority: 3
        timeout: 5s
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
      - name: "myregistry"
        username: foo
        password: bar
        timeout: 20s
        insecureTLS: true
      - name: "docker.io"
        selector: image
        username: foo2
        password: bar2

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
      nomad:
        watchByDefault: true
    ```

## Environment variables

All configuration from file can be transposed into environment variables. As an example, the following configuration:

??? example "diun.yml"
    ```yaml
    db:
      path: diun.db

    watch:
      workers: 10
      schedule: "0 */6 * * *"
      jitter: 30s
      firstCheckNotif: false
      runOnStartup: true

    defaults:
      watchRepo: false
      notifyOn:
        - new
        - update
      sortTags: reverse

    notif:
      gotify:
        endpoint: http://gotify.foo.com
        token: Token123456
        priority: 1
        timeout: 10s
      ntfy:
        endpoint: https://ntfy.sh
        topic: diun-acce65a0-b777-46f9-9a11-58c67d1579c4
        priority: 3
        timeout: 5s
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
      - name: "docker.io"
        selector: image
        username: foo
        password: bar
      - name: "registry.gitlab.com"
        selector: image
        username: fii
        password: bor
        timeout: 20s

    providers:
      kubernetes:
        tlsInsecure: false
        namespaces:
          - default
          - production
    ```

Can be transposed to:

??? example "environment variables"
    ```
    DIUN_DB_PATH=diun.db

    DIUN_WATCH_WORKERS=10
    DIUN_WATCH_SCHEDULE=0 */6 * * *
    DIUN_WATCH_JITTER=30s
    DIUN_WATCH_FIRSTCHECKNOTIF=false
    DIUN_WATCH_RUNONSTARTUP=true

    DIUN_DEFAULTS_WATCHREPO=false
    DIUN_DEFAULTS_NOTIFYON=new,update
    DIUN_DEFAULTS_SORTTAGS=reverse

    DIUN_NOTIF_GOTIFY_ENDPOINT=http://gotify.foo.com
    DIUN_NOTIF_GOTIFY_TOKEN=Token123456
    DIUN_NOTIF_GOTIFY_PRIORITY=1
    DIUN_NOTIF_GOTIFY_TIMEOUT=10s

    DIUN_NOTIF_NTFY_ENDPOINT=https://ntfy.sh
    DIUN_NOTIF_NTFY_TOPIC=diun-acce65a0-b777-46f9-9a11-58c67d1579c4
    DIUN_NOTIF_NTFY_TAGS=whale
    DIUN_NOTIF_NTFY_TIMEOUT=10s

    DIUN_NOTIF_TELEGRAM_TOKEN=aabbccdd:11223344
    DIUN_NOTIF_TELEGRAM_CHATIDS=123456789,987654321

    DIUN_NOTIF_WEBHOOK_ENDPOINT=http://webhook.foo.com/sd54qad89azd5a
    DIUN_NOTIF_WEBHOOK_METHOD=GET
    DIUN_NOTIF_WEBHOOK_HEADERS_CONTENT-TYPE=application/json
    DIUN_NOTIF_WEBHOOK_HEADERS_AUTHORIZATION=Token123456
    DIUN_NOTIF_WEBHOOK_TIMEOUT=10s

    DIUN_REGOPTS_0_NAME=docker.io
    DIUN_REGOPTS_0_SELECTOR=image
    DIUN_REGOPTS_0_USERNAME=foo
    DIUN_REGOPTS_0_PASSWORD=bar
    DIUN_REGOPTS_1_NAME=registry.gitlab.com
    DIUN_REGOPTS_1_SELECTOR=image
    DIUN_REGOPTS_1_USERNAME=fii
    DIUN_REGOPTS_1_PASSWORD=bor
    DIUN_REGOPTS_1_TIMEOUT=20s

    PROVIDERS_KUBERNETES_TLSINSECURE=false
    PROVIDERS_KUBERNETES_NAMESPACES=default,production
    ```

## Reference

* [db](db.md)
* [watch](watch.md)
* [defaults](defaults.md)
* notif
    * [amqp](../notif/amqp.md)
    * [discord](../notif/discord.md)
    * [elasticsearch](../notif/elasticsearch.md)
    * [gotify](../notif/gotify.md)
    * [mail](../notif/mail.md)
    * [matrix](../notif/matrix.md)
    * [mqtt](../notif/mqtt.md)
    * [ntfy](../notif/ntfy.md)
    * [pushover](../notif/pushover.md)
    * [rocketchat](../notif/rocketchat.md)
    * [script](../notif/script.md)
    * [slack](../notif/slack.md)
    * [signal-rest](../notif/signalrest.md)
    * [teams](../notif/teams.md)
    * [telegram](../notif/telegram.md)
    * [webhook](../notif/webhook.md)
* [regopts](regopts.md)
* providers
    * [docker](../providers/docker.md)
    * [kubernetes](../providers/kubernetes.md)
    * [swarm](../providers/swarm.md)
    * [nomad](../providers/nomad.md)
    * [dockerfile](../providers/dockerfile.md)
    * [file](../providers/file.md)
