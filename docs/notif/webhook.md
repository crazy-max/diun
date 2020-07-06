# Webhook notifications

You can send webhook notifications with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      webhook:
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          content-type: application/json
          authorization: Token123456
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `DIUN_NOTIF_WEBHOOK_ENDPOINT`
    * `DIUN_NOTIF_WEBHOOK_METHOD`
    * `DIUN_NOTIF_WEBHOOK_HEADERS_<KEY>`
    * `DIUN_NOTIF_WEBHOOK_TIMEOUT`

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `endpoint`[^1]     |               | URL of the HTTP request |
| `method`[^1]       | `GET`         | HTTP method |
| `headers`          |               | Map of additional headers to be sent (key is case insensitive) |
| `timeout`          | `10s`         | Timeout specifies a time limit for the request to be made |

## Sample

The JSON response will look like this:

```json
{
  "diun_version": "4.0.0",
  "hostname": "myserver",
  "status": "new",
  "provider": "file",
  "image": "docker.io/crazymax/diun:latest",
  "hub_link": "https://hub.docker.com/r/crazymax/diun",
  "mime_type": "application/vnd.docker.distribution.manifest.list.v2+json",
  "digest": "sha256:216e3ae7de4ca8b553eb11ef7abda00651e79e537e85c46108284e5e91673e01",
  "created": "2020-03-26T12:23:56Z",
  "platform": "linux/amd64"
}
```

[^1]: Value required
