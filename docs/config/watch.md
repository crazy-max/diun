# Watch configuration

## `workers`

Maximum number of workers that will execute tasks concurrently. (default `10`)

!!! example "Config file"
    ```yaml
    watch:
      workers: 10
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_WORKERS`

## `schedule`

[CRON expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) to schedule Diun watcher. (default `0 * * * *`)

!!! example "Config file"
    ```yaml
    watch:
      schedule: "0 * * * *"
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_SCHEDULE`

## `firstCheckNotif`

Send notification at the very first analysis of an image. (default `false`)

!!! example "Config file"
    ```yaml
    watch:
      firstCheckNotif: false
    ```

!!! abstract "Environment variables"
    * `DIUN_WATCH_FIRSTCHECKNOTIF`
