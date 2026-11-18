# Matrix notifications

Allow sending notifications to your Matrix server.

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

!!! example "File (access token)"
    ```yaml
    notif:
      matrix:
        homeserverURL: https://matrix.org
        token: myAccessToken
        roomID: "!abcdefGHIjklmno:matrix.org"
    ```

!!! note
    Using `token` or `tokenFile` skips password authentication and does not
    require `user`. Diun does not log out when using a token, so it remains
    valid for subsequent notifications.

| Name               | Default                            | Description                                                                                                                                  |
|--------------------|------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `homeserverURL`    | `https://matrix.org`               | Matrix server URL                                                                                                                            |
| `user`             |                                    | Username for password authentication                                                                                                         |
| `userFile`         |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as username authentication if `user` not defined     |
| `password`         |                                    | Password for authentication                                                                                                                  |
| `passwordFile`     |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as password authentication if `password` not defined |
| `token`            |                                    | Access token. Takes precedence over username and password authentication                                                                     |
| `tokenFile`        |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as access token if `token` not defined               |
| `roomID`           |                                    | Room ID to send messages                                                                                                                     |
| `msgType`          | `notice`                           | Type of message being sent. Can be `notice` or `text`                                                                                        |
| `templateBody`[^1] | See [below](#default-templatebody) | [Notification template](../faq.md#notification-template) for message body                                                                    |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_MATRIX_HOMESERVERURL`
    * `DIUN_NOTIF_MATRIX_USER`
    * `DIUN_NOTIF_MATRIX_USERFILE`
    * `DIUN_NOTIF_MATRIX_PASSWORD`
    * `DIUN_NOTIF_MATRIX_PASSWORDFILE`
    * `DIUN_NOTIF_MATRIX_TOKEN`
    * `DIUN_NOTIF_MATRIX_TOKENFILE`
    * `DIUN_NOTIF_MATRIX_ROOMID`
    * `DIUN_NOTIF_MATRIX_MSGTYPE`
    * `DIUN_NOTIF_MATRIX_TEMPLATEBODY`

### Default `templateBody`

```
[[ config.extra.template.notif.defaultBody ]]
```

## Sample

![](../assets/notif/matrix.png)

[^1]: Value required
