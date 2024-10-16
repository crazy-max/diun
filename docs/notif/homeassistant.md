# MQTT notifications

You can send notifications to any HomeAssistant compatible MQTT broker with the following settings.
The notifier use the auto-discovery specs to create a new sensor per image monitored
A message is sent on each run, even if there is no update or new image, this to allow true monitoring in HA

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

| Name               | Default     | Description                                                            |
|--------------------|-------------|------------------------------------------------------------------------|
| `scheme`[^1]       | `mqtt`      | MQTT server scheme (`mqtt`, `mqtts`, `ws` or `wss`)                    |
| `host`[^1]         | `localhost` | MQTT server host                                                       |
| `port`[^1]         | `1883`      | MQTT server port                                                       |
| `username`         |             | MQTT username                                                          |
| `usernameFile`     |             | Use content of secret file as MQTT username if `username` not defined  |
| `password`         |             | MQTT password                                                          |
| `passwordFile`     |             | Use content of secret file as MQTT password if `password` not defined  |
| `client`[^1]       |             | Client id to be used by this client when connecting to the MQTT broker |
| `discoveryPrefix`  | `homeassistant` | Prefix for the discovery topic                                     |
| `component`[^1]    | `sensor`    | Type of MQTT integration (e.g., `sensor`, `binary_sensor`)             |
| `nodename`[^1]     | `diun`      | Node Name in HA (e.g, `diun`, `docker-hostname`                        |
| `qos`              | `0`         | Ensured message delivery at specified Quality of Service (QoS)         |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_HOMEASSISTANT_SCHEME`
    * `DIUN_NOTIF_HOMEASSISTANT_HOST`
    * `DIUN_NOTIF_HOMEASSISTANT_PORT`
    * `DIUN_NOTIF_HOMEASSISTANT_USERNAME`
    * `DIUN_NOTIF_HOMEASSISTANT_USERNAMEFILE`
    * `DIUN_NOTIF_HOMEASSISTANT_PASSWORD`
    * `DIUN_NOTIF_HOMEASSISTANT_PASSWORDFILE`
    * `DIUN_NOTIF_HOMEASSISTANT_CLIENT`
    * `DIUN_NOTIF_HOMEASSISTANT_DISCOVERYPREFIX`
    * `DIUN_NOTIF_HOMEASSISTANT_COMPONENT`
    * `DIUN_NOTIF_HOMEASSISTANT_QOS`

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
