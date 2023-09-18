# Watch configuration

## Overview

```yaml
watch:
  workers: 10
  schedule: "0 */6 * * *"
  jitter: 30s
  firstCheckNotif: false
  runOnStartup: true
  compareDigest: true
  healthchecks:
    baseURL: https://hc-ping.com/
    uuid: 5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278
  imageDefaults:
    platform:
      os: linux
      arch: amd64
    regopt: ""
    notify_on: [new, update]
    max_tags: 10
    sort_tags: reverse
    include_tags: [latest]
    exclude_tags: [dev]
```

## Configuration

### `workers`

Maximum number of workers that will execute tasks concurrently. (default `10`)

!!! example "Config file"
    ```yaml
    watch:
      workers: 10
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_WORKERS`

### `schedule`

[CRON expression](https://pkg.go.dev/github.com/crazy-max/cron/v3#hdr-CRON_Expression_Format) to schedule Diun.

!!! warning
    Remove this setting if you want to run Diun directly.

!!! example "Config file"
    ```yaml
    watch:
      schedule: "0 */6 * * *"
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_SCHEDULE`

### `jitter`

Enable time jitter. Prior to executing of a job, cron will sleep a random
duration in the range from 0 to _jitter_. (default `30s`)

!!! note
    Only works with `schedule` setting. `0` disables time jitter.

!!! example "Config file"
    ```yaml
    watch:
      schedule: "0 */6 * * *"
      jitter: 30s
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_JITTER`

### `firstCheckNotif`

Send notification at the very first analysis of an image. (default `false`)

!!! example "Config file"
    ```yaml
    watch:
      firstCheckNotif: false
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_FIRSTCHECKNOTIF`

### `runOnStartup`

Check for updates on startup. (default `true`)

!!! example "Config file"
    ```yaml
    watch:
      runOnStartup: true
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_RUNONSTARTUP`

### `compareDigest`

Compare the digest of an image with the registry before downloading the image manifest. It is strongly
recommended leaving this value at `true`, especially with [Docker Hub which imposes a rate-limit](../faq.md#docker-hub-rate-limits)
on image pull. (default `true`)

!!! example "Config file"
    ```yaml
    watch:
      compareDigest: true
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_COMPAREDIGEST`

### `healthchecks`

Healthchecks allows monitoring Diun watcher by sending start and success notification
events to [healthchecks.io](https://healthchecks.io/).

!!! tip
    A [Docker image for Healthchecks](https://github.com/crazy-max/docker-healthchecks) is available if you want
    to self-host your instance.

![](../assets/watch/healthchecks.png)

!!! example "Config file"
    ```yaml
    watch:
      healthchecks:
        baseURL: https://hc-ping.com/
        uuid: 5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_HEALTHCHECKS_BASEURL`
    * `DIUN_WATCH_HEALTHCHECKS_UUID`

* `baseURL`: Base URL for the Healthchecks Ping API (default `https://hc-ping.com/`).
* `uuid`: UUID of an existing healthcheck (required).

### `imageDefaults`

ImageDefaults allows specifying default values for any configuration that is typically set at an Image level. For details More details on these examples can be seen in the [file provider documentation](../providers/file.md). Any value sset at the Image level will override or be merged with any deault values.

!!! tip
    Not all values must be provided. You may chose which ones you'd like to set. A an example may be to specify defaults such that new SemVer tags will be trigger notifications.

!!! tip
    Most values will be strictly overwritten by Image level variables. The notable exception is `metadata`. There are several sources of metadata that can be provided. First, via `ImageDefaults`, second via the platform, and finally by the image. When these are merged, unique keys will always persist but values will be overwritten in the order previously described. Eg. default keys will be overwritten by provider metadata keys on collision.

!!! example "Config file" watching for new x.y.z semver tags
    ```yaml
    watch:
      imageDefaults:
        watch_repo: true
        sort_tags: semver
        include_tags:
          - "^\d+\.\d+\.\d+$"

!!! abstract "Environment variables"
    * `DIUN_WATCH_IMAGE_DEFAULTS_WATCH_REPO`
    * `DIUN_WATCH_IMAGE_DEFAULTS_SORT_TAGS`
    * `DIUN_WATCH_IMAGE_DEFAULTS_INCLUDE_TAGS`
