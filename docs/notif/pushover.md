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
    * `DIUN_NOTIF_PUSHOVER_TOKENFILE`
    * `DIUN_NOTIF_PUSHOVER_RECIPIENT`
    * `DIUN_NOTIF_PUSHOVER_RECIPIENTFILE`

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `token`            |               | Pushover [application/API token](https://pushover.net/api#registration) |
| `tokenFile`        |               | Use content of secret file as Pushover application/API token if `token` not defined |
| `recipient`        |               | User key to send notification to |
| `recipientFile`    |               | Use content of secret file as User key if `recipient` not defined |

## Sample

![](../assets/notif/pushover.png)
