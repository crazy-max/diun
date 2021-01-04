# Pushover notifications

You can send notifications using [Pushover](https://pushover.net/).

## Configuration

!!! example "File"
    ```yaml
    notif:
      pushover:
        token: uQiRzpo4DXghDmr9QzzfQu27cmVRsG
        recipient: gznej3rKEVAvPUxu9vvNnqpmZpokzF
    ```

!!! abstract "Environment variables"
    * `DIUN_NOTIF_PUSHOVER_TOKEN`
    * `DIUN_NOTIF_PUSHOVER_RECIPIENT`

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `token`[^1]        |               | Pushover [application/API token](https://pushover.net/api#registration) |
| `recipient`[^1]    |               | User key to send notification to |

## Sample

![](../assets/notif/pushover.png)

[^1]: Value required
