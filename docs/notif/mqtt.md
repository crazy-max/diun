# Mqtt notifications

You can send notifications to any mqtt compatible server with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      mqtt:
        host: localhost
        port: 1883
        username: guest
        password: guest
        topic: docker/diun
        client: diun
        qos: 0
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `host`[^1]         | `localhost`   | MQTT server host |
| `port`[^1]         | `1883`        | MQTT server port |
| `client`[^1]       | `diun-client` | Name of the client which connects to the server |
| `topic`[^1]        | `docker/diun` | Topic the message will be sent to |
| `username`         |               | MQTT username |
| `usernameFile`     |               | Use content of secret file as MQTT username if `username` not defined |
| `password`         |               | MQTT password |
| `passwordFile`     |               | Use content of secret file as MQTT password if `password` not defined |
| `qos`              | `0`           | Topic the message will be sent to |

## Sample

The JSON response will look like this:

```json
{
  "diun_version": "0.3.0",
  "hostname": "myserver",
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
