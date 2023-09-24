# Script notifications

You can call a script when a notification occurred. Following environment variables will be passed:

```
DIUN_VERSION=4.24.0
DIUN_ENTRY_STATUS=new
DIUN_HOSTNAME=myserver
DIUN_ENTRY_PROVIDER=file
DIUN_ENTRY_IMAGE=docker.io/crazymax/diun:latest
DIUN_ENTRY_HUBLINK=https://hub.docker.com/r/crazymax/diun
DIUN_ENTRY_MIMETYPE=application/vnd.docker.distribution.manifest.list.v2+json
DIUN_ENTRY_DIGEST=sha256:216e3ae7de4ca8b553eb11ef7abda00651e79e537e85c46108284e5e91673e01
DIUN_ENTRY_CREATED=2020-03-26 12:23:56 +0000 UTC
DIUN_ENTRY_PLATFORM=linux/amd64
DIUN_ENTRY_METADATA_CTN_COMMAND=diun serve
DIUN_ENTRY_METADATA_CTN_CREATEDAT=2022-12-29 10:46:20 +0100 CET
DIUN_ENTRY_METADATA_CTN_ID=7c71187fad11aa06f951dee0ebd6382ee0030a8228929fc7ea2fccc18f940788
DIUN_ENTRY_METADATA_CTN_NAMES=diun
DIUN_ENTRY_METADATA_CTN_SIZE=0B
DIUN_ENTRY_METADATA_CTN_STATE=running
DIUN_ENTRY_METADATA_CTN_STATUS=Up Less than a second (health: starting)
```

## Configuration

!!! example "File"
    ```yaml
    notif:
      script:
        cmd: "myprogram"
        args:
          - "--anarg"
          - "another"
    ```

| Name      | Default | Description                                    |
|-----------|---------|------------------------------------------------|
| `cmd`[^1] |         | Command or script to execute                   |
| `args`    |         | List of args to pass to `cmd`                  |
| `dir`     |         | Specifies the working directory of the command |

!!! abstract "Environment variables"
    * `DIUN_NOTIF_SCRIPT_CMD`
    * `DIUN_NOTIF_SCRIPT_ARGS` (comma separated)
    * `DIUN_NOTIF_SCRIPT_DIR`

[^1]: Value required
