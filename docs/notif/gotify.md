# Gotify notifications

Notifications can be sent using a [Gotify](https://gotify.net/) instance.

## Configuration

!!! example "File"
    ```yaml
    notif:
      gotify:
        endpoint: http://gotify.foo.com
        token: Token123456
        priority: 1
        timeout: 10s
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `endpoint`[^1]     |               | Gotify base URL |
| `token`[^1]        |               | Application token |
| `priority`         | `1`           | The priority of the message |
| `timeout`          | `10s`         | Timeout specifies a time limit for the request to be made |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_GOTIFY_ENDPOINT`
    * `DIUN_NOTIF_GOTIFY_TOKEN`
    * `DIUN_NOTIF_GOTIFY_PRIORITY`
    * `DIUN_NOTIF_GOTIFY_TIMEOUT`

## Sample

![](../assets/notif/gotify.png)

[^1]: Value required
