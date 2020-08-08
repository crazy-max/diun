# Rocket.Chat notifications

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
    ```

| Name                  | Default                | Description       |
|-----------------------|------------------------|-------------------|
| `homeserverURL`       | `https://matrix.org`   | Matrix server URL |
| `user`                |                        | Username for authentication |
| `userFile`            |                        | Use content of secret file as username authentication if `username` not defined |
| `password`            |                        | Password for authentication |
| `passwordFile`        |                        | Use content of secret file as password authentication if `password` not defined |
| `roomID`              |                        | Room ID to send messages |
| `msgType`             | `notice`               | Type of message being sent. Can be `notice` or `text` |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_MATRIX_HOMESERVERURL`
    * `DIUN_NOTIF_MATRIX_USER`
    * `DIUN_NOTIF_MATRIX_USERFILE`
    * `DIUN_NOTIF_MATRIX_PASSWORD`
    * `DIUN_NOTIF_MATRIX_PASSWORDFILE`
    * `DIUN_NOTIF_MATRIX_ROOMID`
    * `DIUN_NOTIF_MATRIX_MSGTYPE`

## Sample

![](../assets/notif/matrix.png)
