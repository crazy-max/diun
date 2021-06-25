# Matrix notifications

Allow to send notifications to your Matrix server.

## Configuration

!!! example "File"
    ```yaml
    notif:
      matrix:
        homeserverURL: https://matrix.org
        user: "@foo:matrix.org"
        password: bar
        roomID: "!abcdefGHIjklmno:matrix.org"
        msgType: notice
        templateBody: |
          Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

| Name                  | Default                                    | Description       |
|-----------------------|--------------------------------------------|-------------------|
| `homeserverURL`       | `https://matrix.org`                       | Matrix server URL |
| `user`                |                                            | Username for authentication |
| `userFile`            |                                            | Use content of secret file as username authentication if `username` not defined |
| `password`            |                                            | Password for authentication |
| `passwordFile`        |                                            | Use content of secret file as password authentication if `password` not defined |
| `roomID`              |                                            | Room ID to send messages |
| `msgType`             | `notice`                                   | Type of message being sent. Can be `notice` or `text` |
| `templateBody`[^1]    | See [below](#default-templatebody)         | [Notification template](../faq.md#notification-template) for message body |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_MATRIX_HOMESERVERURL`
    * `DIUN_NOTIF_MATRIX_USER`
    * `DIUN_NOTIF_MATRIX_USERFILE`
    * `DIUN_NOTIF_MATRIX_PASSWORD`
    * `DIUN_NOTIF_MATRIX_PASSWORDFILE`
    * `DIUN_NOTIF_MATRIX_ROOMID`
    * `DIUN_NOTIF_MATRIX_MSGTYPE`
    * `DIUN_NOTIF_MATRIX_TEMPLATEBODY`

### Default `templateBody`

```
[[ config.extra.template.defaultBody ]]
```

## Sample

![](../assets/notif/matrix.png)
