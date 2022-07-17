# Dockerfile provider

## About

The Dockerfile provider allows to parse a [Dockerfile](https://docs.docker.com/engine/reference/builder/) and extract
images for the following instructions:

* [`FROM <image>`](https://docs.docker.com/engine/reference/builder/#from)
* [`COPY --from=<image>`](https://docs.docker.com/engine/reference/builder/#copy)
* [`RUN --mount=type=bind,from=<image>`](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md#run---mounttypebind-the-default-mount-type)

## Quick start

First you have to register the dockerfile provider:

```yaml
db:
  path: diun.db

watch:
  workers: 20
  schedule: "0 */6 * * *"

regopts:
  - name: "docker.io"
    selector: image
    username: foo
    password: bar

providers:
  dockerfile:
    patterns:
      - ./Dockerfile
```

```Dockerfile
# syntax=docker/dockerfile:1.2

# diun.platform=linux/amd64
FROM alpine:latest

# diun.watch_repo=true
# diun.max_tags=10
# diun.platform=linux/amd64
COPY --from=crazymax/yasu / /

# diun.watch_repo=true
# diun.include_tags=^\d+\.\d+\.\d+$
# diun.platform=linux/amd64
RUN --mount=type=bind,target=.,rw \
  --mount=type=bind,from=crazymax/docker:20.10.6,source=/usr/local/bin/docker,target=/usr/bin/docker \
  yasu --version
```

With this Dockerfile the following images will be analyzed:

* `alpine:latest` tag (`linux/amd64` platform)
* Most recent 10 tags for `crazymax/yasu` image (`linux/amd64` platform)
* `crazymax/docker` tags matching `^\d+\.\d+\.\d+$` (`linux/amd64` platform)

Now let's start Diun:

```
$ diun serve --config /etc/diun/diun.yml
Thu, 29 Apr 2021 14:41:55 CEST INF Starting Diun version=4.16.0
Thu, 29 Apr 2021 14:41:55 CEST INF Configuration loaded from file: /etc/diun/diun.yml
Thu, 29 Apr 2021 14:41:55 CEST WRN No notifier available
Thu, 29 Apr 2021 14:41:55 CEST INF Cron triggered
Thu, 29 Apr 2021 14:41:55 CEST INF Found 3 image(s) to analyze provider=dockerfile
Thu, 29 Apr 2021 14:41:59 CEST INF New image found image=docker.io/library/alpine:latest provider=dockerfile
Thu, 29 Apr 2021 14:41:59 CEST INF New image found image=docker.io/crazymax/yasu:latest provider=dockerfile
Thu, 29 Apr 2021 14:42:00 CEST INF New image found image=docker.io/crazymax/yasu:1.14.1 provider=dockerfile
Thu, 29 Apr 2021 14:42:00 CEST INF New image found image=docker.io/crazymax/docker:20.10.6 provider=dockerfile
Thu, 29 Apr 2021 14:42:00 CEST INF New image found image=docker.io/crazymax/yasu:edge provider=dockerfile
Thu, 29 Apr 2021 14:42:01 CEST INF New image found image=docker.io/crazymax/yasu:1.14.0 provider=dockerfile
Thu, 29 Apr 2021 14:42:02 CEST INF New image found image=docker.io/crazymax/docker:20.10.5 provider=dockerfile
Thu, 29 Apr 2021 14:42:02 CEST INF New image found image=docker.io/crazymax/docker:20.10.4 provider=dockerfile
Thu, 29 Apr 2021 14:42:02 CEST INF New image found image=docker.io/crazymax/docker:20.10.3 provider=dockerfile
Thu, 29 Apr 2021 14:42:02 CEST INF New image found image=docker.io/crazymax/docker:20.10.2 provider=dockerfile
Thu, 29 Apr 2021 14:42:03 CEST INF New image found image=docker.io/crazymax/docker:20.10.1 provider=dockerfile
Thu, 29 Apr 2021 14:42:03 CEST INF New image found image=docker.io/crazymax/docker:19.03.15 provider=dockerfile
Thu, 29 Apr 2021 14:42:04 CEST INF New image found image=docker.io/crazymax/docker:19.03.14 provider=dockerfile
Thu, 29 Apr 2021 14:42:04 CEST INF Jobs completed added=13 failed=0 skipped=0 unchanged=0 updated=0
Thu, 29 Apr 2021 14:42:05 CEST INF Cron initialized with schedule 0 */6 * * *
Thu, 29 Apr 2021 14:42:05 CEST INF Next run in 3 hours (2021-04-29 18:00:00 +0200 CEST)
```

## Configuration

### `patterns`

List of path patterns with [matching and globbing supporting patterns](https://github.com/bmatcuk/doublestar/tree/v3).

!!! example "File"
    ```yaml
    providers:
      dockerfile:
        patterns:
          - "**/Dockerfile*"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_DOCKERFILE_PATTERNS` (comma separated)

## Annotations

The following annotations can be added as comments before the target instruction to customize the image analysis:

| Name                | Default      | Description                                                                                                                                                |
|---------------------|--------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `diun.regopt`       |              | [Registry options](../config/regopts.md) name to use                                                                                                       |
| `diun.watch_repo`   | `false`      | Watch all tags of this image                                                                                                                               |
| `diun.notify_on`    | `new;update` | Semicolon separated list of status to be notified: `new`, `update`                                                                                         |
| `diun.sort_tags`    | `reverse`    | [Sort tags method](../faq.md#tags-sorting-when-using-watch_repo) if `diun.watch_repo` enabled. One of `default`, `reverse`, `numerical`, `lexicographical` |
| `diun.max_tags`     | `0`          | Maximum number of tags to watch if `watch_repo` enabled. `0` means all of them                                                                             |
| `diun.include_tags` |              | Semicolon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`                                             |
| `diun.exclude_tags` |              | Semicolon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`                                             |
| `diun.hub_link`     | _automatic_  | Set registry hub link for this image                                                                                                                       |
| `diun.platform`     | _automatic_  | Platform to use (e.g. `linux/amd64`)                                                                                                                       |
