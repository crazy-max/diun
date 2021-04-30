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
| `token`            |               | Telegram bot token (you get your token over the Telegram User @BotFather > `/mybots` > "API Token") |
| `tokenFile`        |               | Use content of secret file as Telegram bot token if `token` not defined |
| `chatIDs`          |               | List of chat IDs to send notifications to (your chat IDs you get if you go to your BOT and send a message to `@getidsbot` `ID: 123456789`) |
| `chatIDsFile`      |               | Use content of secret file as chat IDs if `chatIDs` not defined |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_TELEGRAM_TOKEN`
    * `DIUN_NOTIF_TELEGRAM_TOKENFILE`
    * `DIUN_NOTIF_TELEGRAM_CHATIDS` (comma separated)
    * `DIUN_NOTIF_TELEGRAM_CHATIDSFILE`

!!! example "chat IDs secret file"
    Chat IDs secret file must be a valid JSON array like: `[123456789,987654321]`

## Sample

![](../assets/notif/telegram.png)

[^1]: Value required
