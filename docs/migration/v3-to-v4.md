# Diun v3 to v4

## Timeout value as duration

Only accept duration as timeout value (`10` becomes `10s`)

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
      schedule: "0 * * * *"
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

??? example "v4"
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
