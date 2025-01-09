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

| Name           | Default | Description                                                    |
|----------------|---------|----------------------------------------------------------------|
| `endpoint`[^1] |         | URL of the HTTP request                                        |
| `method`[^1]   | `GET`   | HTTP method                                                    |
| `headers`      |         | Map of additional headers to be sent (key is case insensitive) |
| `timeout`      | `10s`   | Timeout specifies a time limit for the request to be made      |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_WEBHOOK_ENDPOINT`
    * `DIUN_NOTIF_WEBHOOK_METHOD`
    * `DIUN_NOTIF_WEBHOOK_HEADERS_<KEY>`
    * `DIUN_NOTIF_WEBHOOK_TIMEOUT`

## Sample

The JSON request will look like this:

```json
{
  "diun_version": "4.24.0",
  "hostname": "myserver",
  "status": "new",
  "provider": "file",
  "image": "docker.io/crazymax/diun:latest",
  "hub_link": "https://hub.docker.com/r/crazymax/diun",
  "mime_type": "application/vnd.docker.distribution.manifest.list.v2+json",
  "digest": "sha256:216e3ae7de4ca8b553eb11ef7abda00651e79e537e85c46108284e5e91673e01",
  "created": "2020-03-26T12:23:56Z",
  "platform": "linux/amd64",
  "metadata": {
    "ctn_command": "diun serve",
    "ctn_createdat": "2022-12-29 10:22:15 +0100 CET",
    "ctn_id": "0dbd10e15b31add2c48856fd34451adabf50d276efa466fe19a8ef5fbd87ad7c",
    "ctn_names": "diun",
    "ctn_size": "0B",
    "ctn_state": "running",
    "ctn_status": "Up Less than a second (health: starting)"
  }
}
```

[^1]: Value required
