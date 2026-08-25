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
        user: "@diun:matrix.org"
        accessToken: syt_dGhpc19pc19hX2Zha2VfdG9rZW4K
        roomID: "!abcdefGHIjklmno:matrix.org"
        msgType: notice
        templateBody: |
          Docker tag {{ .Entry.Image }} which you subscribed to through {{ .Entry.Provider }} provider has been released.
    ```

    When `accessToken` (or `accessTokenFile`) is set, diun authenticates by presenting the bearer token on every request and does not call the `/login` or `/logout` endpoints. The token is reused across notifications, so it should be treated as a long-lived secret.

| Name               | Default                            | Description                                                                                                                                  |
|--------------------|------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `homeserverURL`    | `https://matrix.org`               | Matrix server URL                                                                                                                            |
| `user`             |                                    | Username for authentication. Required.                                                                                                       |
| `userFile`         |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as username authentication if `username` not defined |
| `accessToken`      |                                    | Long-lived access token for authentication. Mutually exclusive with `password`                                                                |
| `accessTokenFile`  |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as access token authentication if `accessToken` not defined |
| `password`         |                                    | Password for authentication. Mutually exclusive with `accessToken`                                                                            |
| `passwordFile`     |                                    | Use content of [secret file](../faq.md#secrets-loaded-from-files-and-trailing-newlines) as password authentication if `password` not defined |
| `roomID`           |                                    | Room ID to send messages                                                                                                                     |
| `msgType`          | `notice`                           | Type of message being sent. Can be `notice` or `text`                                                                                        |
| `templateBody`[^1] | See [below](#default-templatebody) | [Notification template](../faq.md#notification-template) for message body                                                                    |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_MATRIX_HOMESERVERURL`
    * `DIUN_NOTIF_MATRIX_USER`
    * `DIUN_NOTIF_MATRIX_USERFILE`
    * `DIUN_NOTIF_MATRIX_PASSWORD`
    * `DIUN_NOTIF_MATRIX_PASSWORDFILE`
    * `DIUN_NOTIF_MATRIX_ACCESSTOKEN`
    * `DIUN_NOTIF_MATRIX_ACCESSTOKENFILE`
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
