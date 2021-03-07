# MQTT notifications

You can send notifications to any MQTT compatible server with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      mqtt:
        scheme: mqtt
        host: localhost
        port: 1883
        username: guest
        password: guest
        client: diun
        topic: docker/diun
        qos: 0
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `scheme`[^1]       | `mqtt`        | MQTT server scheme (`mqtt`, `mqtts`, `ws` or `wss`) |
| `host`[^1]         | `localhost`   | MQTT server host |
| `port`[^1]         | `1883`        | MQTT server port |
| `username`         |               | MQTT username |
| `usernameFile`     |               | Use content of secret file as MQTT username if `username` not defined |
| `password`         |               | MQTT password |
| `passwordFile`     |               | Use content of secret file as MQTT password if `password` not defined |
| `client`[^1]       |               | Client id to be used by this client when connecting to the MQTT broker |
| `topic`[^1]        |               | Topic the message will be sent to |
| `qos`              | `0`           | Ensured message delivery at specified Quality of Service (QoS) |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_MQTT_SCHEME`
    * `DIUN_NOTIF_MQTT_HOST`
    * `DIUN_NOTIF_MQTT_PORT`
    * `DIUN_NOTIF_MQTT_USERNAME`
    * `DIUN_NOTIF_MQTT_USERNAMEFILE`
    * `DIUN_NOTIF_MQTT_PASSWORD`
    * `DIUN_NOTIF_MQTT_PASSWORDFILE`
    * `DIUN_NOTIF_MQTT_CLIENT`
    * `DIUN_NOTIF_MQTT_TOPIC`
    * `DIUN_NOTIF_MQTT_QOS`

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
