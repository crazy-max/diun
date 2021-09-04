# Teams notifications

You can send notifications to your Teams team-channel using an [incoming webhook URL](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/what-are-webhooks-and-connectors).

## Configuration

!!! example "File"
    ```yaml
    notif:
      teams:
        webhookURL: https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij
        renderFacts: true
        templateBody: |
          Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

| Name               | Default                                    | Description   |
|--------------------|--------------------------------------------|---------------|
| `webhookURL`[^1]   |                                            | Teams [incoming webhook URL](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/what-are-webhooks-and-connectors) |
| `renderFacts`      | `true`                                     | Render fact objects |
| `templateBody`[^1] | See [below](#default-templatebody)         | [Notification template](../faq.md#notification-template) for message body |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_TEAMS_WEBHOOKURL`
    * `DIUN_NOTIF_TEAMS_RENDERFACTS`
    * `DIUN_NOTIF_TEAMS_TEMPLATEBODY`

### Default `templateBody`

```
Docker tag {{ if .Entry.Image.HubLink }}[`{{ .Entry.Image }}`]({{ .Entry.Image.HubLink }}){{ else }}`{{ .Entry.Image }}`{{ end }}{{ if (eq .Entry.Status "new") }}newly added{{ else }}updated{{ end }}.
```

## Sample

![](../assets/notif/teams.png)

[^1]: Value required
