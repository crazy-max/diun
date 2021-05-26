# Diun v3 to v4

## Timeout value as duration

Only accept duration as timeout value (`10` becomes `10s`)

## Registry options enhancements

Configuration of registry options has changed:

??? example "v3"
    ```yaml
    regopts:
      myregistry:
        username: fii
        password: bor
        insecure_tls: true
        timeout: 5s
      docker.io:
        username: foo
        password: bar
      docker.io/crazymax:
        username_file: /run/secrets/username
        password_file: /run/secrets/password
    ```

??? example "v4"
    ```yaml
    regopts:
      - name: "myregistry"
        username: fii
        password: bor
        insecureTLS: true
        timeout: 5s
      - name: "docker.io"
        selector: image
        username: foo
        password: bar
      - name: "docker.io/crazymax"
        selector: image
        usernameFile: /run/secrets/username
        passwordFile: /run/secrets/password
    ```

Also, registry options can now be resolved automatically based on image name.
Take a look at the [Registry options configuration](../config/regopts.md) for more details.

## Configuration transposed into environment variables

All configuration is now transposed into environment variables. Take a look at the [documentation](../config/index.md#environment-variables) for more details.

`DIUN_DB` env var has been renamed `DIUN_DB_PATH` to follow environment variables transposition.

## All fields in configuration are now _camelCased_

In order to enable transposition into environmental variables, all fields in configuration are now _camelCased_:

* `notif.mail.insecure_skip_verify` > `notif.mail.insecureSkipVerify`
* `notif.rocketchat.user_id` > `notif.rocketchat.userID`
* `watch.first_check_notif` > `watch.firstCheckNotif`
* ...

??? example "v3"
    ```yaml
    db:
      path: diun.db
    
    watch:
      workers: 10
      schedule: "0 */6 * * *"
      first_check_notif: false
    
    notif:
      amqp:
        host: localhost
        port: 5672
        username: guest
        password: guest
        exchange: 
        queue: queue
      gotify:
        endpoint: http://gotify.foo.com
        token: Token123456
        priority: 1
        timeout: 10
      mail:
        host: localhost
        port: 25
        ssl: false
        insecure_skip_verify: false
        username:
        password:
        from:
        to:
      rocketchat:
        endpoint: http://rocket.foo.com:3000
        channel: "#general"
        user_id: abcdEFGH012345678
        token: Token123456
        timeout: 10
      script:
          cmd: "myprogram"
          args:
            - "--anarg"
            - "another"
      slack:
        webhook_url: https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
      teams:
        webhook_url: https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
      telegram:
        token: aabbccdd:11223344
        chat_ids:
          - 123456789
          - 987654321
      webhook:
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          Content-Type: application/json
          Authorization: Token123456
        timeout: 10
    
    regopts:
      myregistry:
        username: fii
        password: bor
        insecure_tls: true
        timeout: 5s
      docker.io:
        username: foo
        password: bar
      docker.io/crazymax:
        username_file: /run/secrets/username
        password_file: /run/secrets/password
    
    providers:
      docker:
        watch_stopped: true
      swarm:
        watch_by_default: true
      file:
        directory: ./imagesdir
    ```

??? example "v4"
    ```yaml
    db:
      path: diun.db
    
    watch:
      workers: 10
      schedule: "0 */6 * * *"
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
      - name: "myregistry"
        username: fii
        password: bor
        insecureTLS: true
        timeout: 5s
      - name: "docker.io"
        selector: image
        username: foo
        password: bar
      - name: "docker.io/crazymax"
        selector: image
        usernameFile: /run/secrets/username
        passwordFile: /run/secrets/password
    
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

## `/diun.yml` not loaded by default in Docker image

Following the transposition of the configuration into environment variables, the configuration file `/diun.yml`
is no longer loaded by default in the official Docker image.

If you want to load a configuration file through the Docker image you will have to declare the
[`CONFIG` environment variable with `serve` command](../usage/command-line.md#serve) pointing to the assigned
configuration file:

!!! tip
    This is no longer required since version 4.2.0. Now configuration file can be loaded from
    [default places](../config/index.md#configuration-file)

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
