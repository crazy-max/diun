# Telegram notifications

Notifications can be sent via Telegram using a [Telegram Bot](https://core.telegram.org/bots).

Follow the [instructions](https://core.telegram.org/bots#6-botfather) to set up a bot and get it's token.

Message the [GetID bot](https://t.me/getidsbot) to find your chat ID.
Multiple chat IDs can be provided in order to deliver notifications to multiple recipients.

## Configuration

!!! example "File"
    ```yaml
    notif:
      telegram:
        token: aabbccdd:11223344
        chatIDs:
          - 123456789
          - 987654321
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `token`[^1]        |               | Telegram bot token |
| `chatIDs`[^1]      |               | List of chat IDs to send notifications to |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_TELEGRAM_TOKEN`
    * `DIUN_NOTIF_TELEGRAM_CHATIDS` (comma separated)

## Sample

![](../assets/notif/telegram.png)

[^1]: Value required
