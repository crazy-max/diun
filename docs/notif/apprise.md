# Apprise notifications

Notifications can be sent using an apprise api instance.

## Configuration

!!! example "File"
    ```yaml
        notif:
          apprise:
            endpoint: http://apprise:8000
            token: abc
            tags:
              - diun
            timeout: 10s
            templateTitle: "{{ .Entry.Image }} released"
            templateBody: |
              Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

| Name            | Default                             | Description                                                                |
|-----------------|-------------------------------------|----------------------------------------------------------------------------|
| `endpoint`[^1]  |                                     | Hostname and port of your apprise api instance                             |
| `token`[^2]     |                                     | token representing your config file (Config Key)                           |
| `tokenFile`     |                                     | Use content of secret file as application token if `token` not defined     |
| `tags`          |                                     | List of Tags in your config file you want to notify                        |
| `urls`[^2]      |                                     | List of [URLs](https://github.com/caronc/apprise/wiki/URLBasics) to notify |
| `timeout`       | `10s`                               | Timeout specifies a time limit for the request to be made                  |
| `templateTitle` | See [below](#default-templatetitle) | [Notification template](../faq.md#notification-template) for message title |
| `templateBody`  | See [below](#default-templatebody)  | [Notification template](../faq.md#notification-template) for message body  |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_APPRISE_ENDPOINT`
    * `DIUN_NOTIF_APPRISE_TOKEN`
    * `DIUN_NOTIF_APPRISE_TAGS`
    * `DIUN_NOTIF_APPRISE_URLS`
    * `DIUN_NOTIF_APPRISE_TIMEOUT`
    * `DIUN_NOTIF_APPRISE_TEMPLATETITLE`
    * `DIUN_NOTIF_APPRISE_TEMPLATEBODY`

### Default `templateTitle`

```
[[ config.extra.template.notif.defaultTitle ]]
```

### Default `templateBody`

```
[[ config.extra.template.notif.defaultBody ]]
```

[^1]: Value required
[^2]: One of these 2 values is required
