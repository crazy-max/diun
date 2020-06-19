# Amqp notifications

You can send notifications to any amqp compatible server with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      amqp:
        host: localhost
        port: 5672
        username: guest
        password: guest
        queue: queue
    ```

!!! abstract "Environment variables"
    * `DIUN_NOTIF_AMQP_HOST`
    * `DIUN_NOTIF_AMQP_EXCHANGE`
    * `DIUN_NOTIF_AMQP_PORT`
    * `DIUN_NOTIF_AMQP_USERNAME`
    * `DIUN_NOTIF_AMQP_USERNAMEFILE`
    * `DIUN_NOTIF_AMQP_PASSWORD`
    * `DIUN_NOTIF_AMQP_PASSWORDFILE`
    * `DIUN_NOTIF_AMQP_QUEUE`

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `host`[^1]         | `localhost`   | AMQP server host |
| `port`[^1]         | `5672`        | AMQP server port |
| `username`         |               | AMQP username |
| `usernameFile`     |               | Use content of secret file as AMQP username if `username` not defined |
| `password`         |               | AMQP password |
| `passwordFile`     |               | Use content of secret file as AMQP password if `password` not defined |
| `exchange`         |               | Name of the exchange the message will be sent to |
| `queue`[^1]        |               | Name of the queue the message will be sent to |

## Sample

The JSON response will look like this:

```json
{
  "diun_version": "0.3.0",
  "status": "new",
  "provider": "file",
  "image": "docker.io/crazymax/swarm-cronjob:0.2.1",
  "hub_link": "https://hub.docker.com/r/crazymax/swarm-cronjob",
  "mime_type": "application/vnd.docker.distribution.manifest.v2+json",
  "digest": "sha256:5913d4b5e8dc15430c2f47f40e43ab2ca7f2b8df5eee5db4d5c42311e08dfb79",
  "created": "2019-01-24T10:26:49.152006005Z",
  "platform": "linux/amd64"
}
```

[^1]: Value required
