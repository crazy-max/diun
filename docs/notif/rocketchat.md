# Rocket.Chat notifications

Allow to send notifications to your Rocket.Chat channel.

## Configuration

!!! example "File"
    ```yaml
    notif:
      rocketchat:
        endpoint: http://rocket.foo.com:3000
        channel: "#general"
        userID: abcdEFGH012345678
        token: Token123456
        timeout: 10s
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `endpoint`[^1]     |               | Rocket.Chat base URL |
| `channel`[^1]      |               | Channel name with the prefix in front of it |
| `userID`[^1]       |               | User ID |
| `token`[^1]        |               | Authentication token |
| `timeout`          | `10s`         | Timeout specifies a time limit for the request to be made |

!!! warning
    You must first create a _Personal Access Token_ through your account settings on your Rocket.Chat instance.

!!! abstract "Environment variables"
    * `DIUN_NOTIF_ROCKETCHAT_ENDPOINT`
    * `DIUN_NOTIF_ROCKETCHAT_CHANNEL`
    * `DIUN_NOTIF_ROCKETCHAT_USERID`
    * `DIUN_NOTIF_ROCKETCHAT_TOKEN`
    * `DIUN_NOTIF_ROCKETCHAT_TIMEOUT`

## Sample

![](../assets/notif/rocketchat.png)

[^1]: Value required
