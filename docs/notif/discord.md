# Discord notifications

Allow to send notifications to your Discord channel.

## Configuration

!!! example "File"
    ```yaml
    notif:
      discord:
        webhookURL: https://discordapp.com/api/webhooks/1234567890/Abcd-eFgh-iJklmNo_pqr
        mentions:
          - "@here"
          - "@everyone"
          - "<@124>"
          - "<@125>"
          - "<@&200>"
        timeout: 10s
        templateTitle: "{{ .Entry.Image }} released"
        templateBody: |
          Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

| Name                | Default                               | Description   |
|---------------------|---------------------------------------|---------------|
| `webhookURL`[^1]    |                                       | Discord [incoming webhook URL](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) |
| `mentions`          |                                       | List of users or roles to notify |
| `timeout`           | `10s`                                 | Timeout specifies a time limit for the request to be made |
| `templateTitle`[^1] | See [below](#default-templatetitle)   | [Notification template](../faq.md#notification-template) for message title |
| `templateBody`[^1]  | See [below](#default-templatebody)    | [Notification template](../faq.md#notification-template) for message body |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_DISCORD_WEBHOOKURL`
    * `DIUN_NOTIF_DISCORD_MENTIONS` (comma separated)
    * `DIUN_NOTIF_DISCORD_TIMEOUT`
    * `DIUN_NOTIF_DISCORD_TEMPLATETITLE`
    * `DIUN_NOTIF_DISCORD_TEMPLATEBODY`

### Default `templateTitle`

```
[[ config.extra.template.defaultTitle ]]
```

### Default `templateBody`

```
[[ config.extra.template.defaultBody ]]
```

## Sample

![](../assets/notif/discord-1.png)

![](../assets/notif/discord-2.png)

[^1]: Value required
