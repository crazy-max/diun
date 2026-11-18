# Prometheus metrics configuration

## Overview

Diun can expose Prometheus metrics from a dedicated HTTP server. This server is
disabled by default.

```yaml
metrics:
  enabled: false
  addr: ":9090"
  path: /metrics
  tokenFile: /run/secrets/diun_metrics_token
```

!!! warning
    The metrics endpoint has no authentication unless `token` or `tokenFile` is
    configured. Bind it to a private interface, avoid exposing it to the
    Internet, and use a TLS reverse proxy if Prometheus scrapes it over an
    untrusted network.

## Configuration

### `enabled`

Enable the Prometheus metrics HTTP server. (default `false`)

!!! example "Config file"
    ```yaml
    metrics:
      enabled: true
    ```

!!! abstract "Environment variables"
    * `DIUN_METRICS_ENABLED`

### `addr`

Address the Prometheus metrics HTTP server listens on. (default `:9090`)

!!! example "Config file"
    ```yaml
    metrics:
      addr: ":9090"
    ```

!!! abstract "Environment variables"
    * `DIUN_METRICS_ADDR`

### `path`

HTTP path used to expose Prometheus metrics. (default `/metrics`)

!!! example "Config file"
    ```yaml
    metrics:
      path: /metrics
    ```

!!! abstract "Environment variables"
    * `DIUN_METRICS_PATH`

### `token`

Bearer token required to scrape the metrics endpoint.

!!! example "Config file"
    ```yaml
    metrics:
      token: very-secret-token
    ```

!!! abstract "Environment variables"
    * `DIUN_METRICS_TOKEN`

### `tokenFile`

Path to a file containing the bearer token required to scrape the metrics
endpoint. If `token` is also set, `token` takes precedence.

!!! example "Config file"
    ```yaml
    metrics:
      tokenFile: /run/secrets/diun_metrics_token
    ```

!!! abstract "Environment variables"
    * `DIUN_METRICS_TOKENFILE`

### `metadataLabels`

Entry metadata keys to expose as additional labels on all per-image metrics.
Empty by default: only explicitly listed keys are exposed, each prefixed with
`metadata_` to avoid collisions with native Diun labels.

!!! example "Config file"
    ```yaml
    metrics:
      enabled: true
      metadataLabels:
        - channel
    ```

    With `metadata.channel: newest` on an image entry, per-image metrics include
    the label `metadata_channel="newest"`:

    ```text
    diun_image_created_timestamp_seconds{provider="file",image="docker.io/example/image:v1.2.3",metadata_channel="newest"}
    ```

    Entries without the selected metadata key expose an empty label value, such
    as `metadata_channel=""`, so the label schema stays consistent.

!!! abstract "Environment variables"
    * `DIUN_METRICS_METADATALABELS` (comma-separated list)

!!! warning
    Additional labels increase metric cardinality, and the selected values are
    part of the series identity: entries sharing a `provider` and `image` but
    differing in the selected metadata are exposed as separate series. Only
    select metadata keys with a small, bounded set of values.

## Prometheus scrape configuration

```yaml
scrape_configs:
  - job_name: diun
    metrics_path: /metrics
    static_configs:
      - targets:
          - diun:9090
```

With bearer authentication enabled:

```yaml
scrape_configs:
  - job_name: diun
    metrics_path: /metrics
    authorization:
      type: Bearer
      credentials_file: /etc/prometheus/secrets/diun_metrics_token
    static_configs:
      - targets:
          - diun:9090
```

## Metrics

Diun exposes Go runtime and process metrics from the Prometheus Go client, plus
the following application metrics:

| Metric                                    | Type    | Labels                        | Description                                                                                                                                                |
| ----------------------------------------- | ------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `diun_build_info`                         | Gauge   | `version`                     | Build information for the Diun instance.                                                                                                                   |
| `diun_watch_runs_total`                   | Counter |                               | Completed watch runs.                                                                                                                                      |
| `diun_watch_skipped_runs_total`           | Counter |                               | Watch runs skipped because another run was already active.                                                                                                 |
| `diun_watch_last_run_timestamp_seconds`   | Gauge   |                               | Unix timestamp of the last completed watch run.                                                                                                            |
| `diun_watch_last_run_duration_seconds`    | Gauge   |                               | Duration in seconds of the last completed watch run.                                                                                                       |
| `diun_watch_last_run_images`              | Gauge   | `status`                      | Number of images by status in the last completed watch run.                                                                                                |
| `diun_image_update_available`             | Gauge   | `provider`, `image`           | `1` when the last check found an actionable update for the image, otherwise `0`. First-run baseline `new` results are not treated as actionable updates.   |
| `diun_image_last_check_timestamp_seconds` | Gauge   | `provider`, `image`           | Unix timestamp of the last completed check for the image.                                                                                                  |
| `diun_image_last_check_status`            | Gauge   | `provider`, `image`, `status` | Last check status for the image. The active status has value `1`.                                                                                          |
| `diun_image_created_timestamp_seconds`    | Gauge   | `provider`, `image`           | Unix timestamp of the image manifest creation time reported by the registry. This metric is omitted if the registry does not provide a creation timestamp. |
| `diun_image_local_digest_mismatch`        | Gauge   | `provider`, `image`           | Runtime digest comparison; see [local comparison](#local-digest-comparison).                                                                               |

By default, the per-image metrics intentionally use only the `provider`,
`image`, and `status` labels to keep cardinality predictable across Docker,
Swarm, Kubernetes, Nomad, Dockerfile, and file providers.
[`metadataLabels`](#metadatalabels) adds the selected prefixed labels to all
per-image metrics.

See [Docker Compose with Prometheus metrics](../faq.md#docker-compose-with-prometheus-metrics)
for a complete Compose example.

### Local digest comparison

`diun_image_local_digest_mismatch` compares watched deployments with the
registry digest recorded by Diun's last watch run. It refreshes on every scrape
without querying the registry:

- `1`: at least one deployed container has a known digest mismatch.
- `0`: all compared containers match.
- Absent: no deployment can be compared, or the result is unknown. This does not
  mean the deployment is up to date.

Pulling and recreating containers to the recorded digest clears the mismatch
without waiting for another watch run. An unchanged registry poll alone does
not clear it. `diun_image_update_available` still describes the last registry
check and is unaffected.

Supported providers honor their existing watch settings:

- **Docker:** compares the images backing watched containers.
- **Swarm:** compares running tasks, including old tasks during rollouts.
  Requires permission to list services and tasks.
- **Kubernetes:** compares regular containers across all selected Pod replicas.
  Uses existing Pod-list permissions; opaque runtime image IDs are unknown.

Provider lookups share a three-second deadline. A failed lookup does not
suppress other providers. Unknown digests omit the series unless a known
mismatch exists. Metadata labels identify the recorded watch entry; the
comparison still includes all watched deployments of that provider and image.

A mismatch does not prove an image is older: it can also mean the deployment is
ahead of Diun's last poll, or a multi-platform index changed without changing
the running platform image.

## Alert example

```yaml
groups:
  - name: diun
    rules:
      - alert: DiunImageUpdateAvailable
        expr: diun_image_update_available == 1
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Image update available"
          description: "{{ $labels.image }} has an update available from the {{ $labels.provider }} provider."
```
