# Signal-REST notifications

The notification uses the [Signal REST API](https://github.com/bbernhard/signal-cli-rest-api).

You can send Signal notifications via the Signal REST API with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      signalrest:
        endpoint: http://192.168.42.50:8080/v2/send
        number: "+00471147111337"
        recipients:
          - "+00472323111337"
        timeout: 10s
    ```

| Name               | Default                         | Description                                               |
|--------------------|---------------------------------|-----------------------------------------------------------|
| `endpoint`         | `http://localhost:8080/v2/send` | URL of the Signal REST API endpoint                       |
| `number`[^1]       |                                 | The senders number you registered                         |
| `recipients`[^1]   |                                 | A list of recipients, either phone numbers or group ID's  |
| `timeout`          | `10s`                           | Timeout specifies a time limit for the request to be made |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_SIGNALREST_ENDPOINT`
    * `DIUN_NOTIF_SIGNALREST_NUMBER`
    * `DIUN_NOTIF_SIGNALREST_RECIPIENTS_<KEY>`
    * `DIUN_NOTIF_SIGNALREST_TIMEOUT`

## Sample

The message you receive in your Signal App will look like this:

```text
Docker tag docker.io/diun/testnotif:latest which you subscribed to through file provider new has been updated on docker.io registry (triggered by5bfaae601770 host).
```

[^1]: Value required
