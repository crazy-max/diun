# Ntfy notifications

Notifications can be sent using a [ntfy](https://ntfy.sh/) instance.

## Configuration

!!! example "File"
    ```yaml
        notif:
          ntfy:
            endpoint: https://ntfy.sh
            topic: diun-acce65a0-b777-46f9-9a11-58c67d1579c4
            priority: 3
            tags:
              - whale
            timeout: 10s
            templateTitle: "{{ .Entry.Image }} released"
            templateBody: |
              Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

| Name                | Default                             | Description                                                                |
| ------------------- | ----------------------------------- | -------------------------------------------------------------------------- |
| `endpoint`[^1]      | `https://ntfy.sh`                   | Ntfy base URL                                                              |
| `topic`             |                                     | Ntfy topic                                                                 |
| `priority`          | 3                          | The priority of the message                                                |
| `tags`              | `["package"]`                       | Emoji to go in your notiication                                            |
| `timeout`           | `10s`                               | Timeout specifies a time limit for the request to be made                  |
| `templateTitle`[^1] | See [below](#default-templatetitle) | [Notification template](../faq.md#notification-template) for message title |
| `templateBody`[^1]  | See [below](#default-templatebody)  | [Notification template](../faq.md#notification-template) for message body  |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_NTFY_ENDPOINT`
    * `DIUN_NOTIF_NTFY_TOPIC`
    * `DIUN_NOTIF_NTFY_PRIORITY`
    * `DIUN_NOTIF_NTFY_TAGS`
    * `DIUN_NOTIF_NTFY_TIMEOUT`
    * `DIUN_NOTIF_NTFY_TEMPLATETITLE`
    * `DIUN_NOTIF_NTFY_TEMPLATEBODY`

### Default `templateTitle`

```
[[ config.extra.template.notif.defaultTitle ]]
```

### Default `templateBody`

```
[[ config.extra.template.notif.defaultBody ]]
```

[^1]: Value required
