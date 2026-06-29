# Containerd provider

## About

The Containerd provider allows you to analyze containers from a containerd
instance through its gRPC API socket to extract images found and check for
updates on the registry.

## Quick start

Here we use a single Containerd provider with a minimum configuration to analyze
labeled containers in the `default` namespace of your local containerd instance.

```yaml
watch:
  workers: 20
  schedule: "0 */6 * * *"

providers:
  containerd: {}
```

If Diun runs from the Docker image, mount the containerd socket:

```yaml
services:
  diun:
    image: crazymax/diun:latest
    command: serve
    volumes:
      - "./data:/data"
      - "/run/containerd/containerd.sock:/run/containerd/containerd.sock"
      - "./diun.yml:/diun.yml:ro"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
    restart: always
```

Then run a labeled container with `nerdctl`:

```shell
nerdctl run -d --name redis --label diun.enable=true redis:6.2.3-alpine
```

## Configuration

!!! hint
    Environment variable `DIUN_PROVIDERS_CONTAINERD=true` can be used to enable this provider with default values.

### `endpoint`

Containerd gRPC socket to connect to. Local containerd socket if empty.

!!! example "File"
    ```yaml
    providers:
      containerd:
        endpoint: "/run/containerd/containerd.sock"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_CONTAINERD_ENDPOINT`

### `namespaces`

Containerd namespaces to query (default `default`).

!!! example "File"
    ```yaml
    providers:
      containerd:
        namespaces:
          - default
          - production
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_CONTAINERD_NAMESPACES`

### `watchByDefault`

Enable watch by default. If false, containers that don't have `diun.enable=true` label will be ignored (default `false`).

!!! example "File"
    ```yaml
    providers:
      containerd:
        watchByDefault: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_CONTAINERD_WATCHBYDEFAULT`

### `watchStopped`

Include stopped containers too (default `false`).

!!! example "File"
    ```yaml
    providers:
      containerd:
        watchStopped: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_CONTAINERD_WATCHSTOPPED`

## Containerd labels

You can configure more finely the way to analyze the image of your container through containerd labels:

| Name                        | Default                        | Description                                                                                                                                             |
|-----------------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `diun.enable`               |                                | Set to true to enable image analysis of this container                                                                                                  |
| `diun.regopt`               |                                | [Registry options](../config/regopts.md) name to use                                                                                                    |
| `diun.watch_repo`           | `false`                        | Watch all tags of this container image ([be careful](../faq.md#docker-hub-rate-limits) with this setting)                                               |
| `diun.watch_newer_only`     | `false`                        | Only notify for tags whose semver is strictly greater than the current tag. Non-semver tags (e.g. `latest`) are ignored. Requires `diun.watch_repo`      |
| `diun.include_prereleases`  | `false`                        | When `diun.watch_newer_only` is enabled, also include pre-release tags (e.g. `-rc.1`, `-alpha`). Requires `diun.watch_newer_only`                       |
| `diun.notify_on`            | `new;update`                   | Semicolon separated list of status to be notified: `new`, `update`                                                                                      |
| `diun.sort_tags`            | `reverse`                      | [Sort tags method](../faq.md#tags-sorting-when-using-watch_repo) if `diun.watch_repo` enabled. One of `default`, `reverse`, `semver`, `lexicographical` |
| `diun.max_tags`             | `0`                            | Maximum number of tags to watch if `diun.watch_repo` enabled. `0` means all of them                                                                     |
| `diun.include_tags`         |                                | Semicolon separated list of regular expressions to include tags. If set, replaces `defaults.includeTags` for this image. Can be useful if you enable `diun.watch_repo` |
| `diun.exclude_tags`         |                                | Semicolon separated list of regular expressions to exclude tags. If set, replaces `defaults.excludeTags` for this image. Can be useful if you enable `diun.watch_repo` |
| `diun.hub_link`             | _automatic_                    | Set registry hub link for this image                                                                                                                    |
| `diun.platform`             | _automatic_                    | Platform to use (e.g. `linux/amd64`)                                                                                                                    |
| `diun.metadata.*`           | See [below](#default-metadata) | Additional metadata that can be used in [notification template](../faq.md#notification-template) (e.g. `diun.metadata.foo=bar`)                         |

## Default metadata

| Key                              | Description          |
|----------------------------------|----------------------|
| `diun.metadata.ctn_id`           | Container ID         |
| `diun.metadata.ctn_name`         | Container name       |
| `diun.metadata.ctn_image`        | Container image      |
| `diun.metadata.ctn_namespace`    | Container namespace  |
| `diun.metadata.ctn_createdat`    | Container created at |
| `diun.metadata.ctn_updatedat`    | Container updated at |
| `diun.metadata.ctn_runtime`      | Container runtime    |
| `diun.metadata.ctn_snapshotter`  | Snapshotter name     |
| `diun.metadata.ctn_snapshot_key` | Snapshot key         |
| `diun.metadata.ctn_status`       | Task status          |
