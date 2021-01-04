# Slack notifications

You can send notifications to your Slack channel using an [incoming webhook URL](https://api.slack.com/messaging/webhooks).

!!! hint
    Mattermost webhooks are compatible with Slack notification without any special configuration (if Webhooks are enabled).

## Configuration

!!! example "File"
    ```yaml
    notif:
      slack:
        webhookURL: https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `webhookURL`[^1]   |               | Slack [incoming webhook URL](https://api.slack.com/messaging/webhooks) |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_SLACK_WEBHOOKURL`

## Sample

![](../assets/notif/slack.png)

[^1]: Value required
