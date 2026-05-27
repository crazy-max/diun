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

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `diun_build_info` | Gauge | `version` | Build information for the Diun instance. |
| `diun_watch_runs_total` | Counter | | Completed watch runs. |
| `diun_watch_skipped_runs_total` | Counter | | Watch runs skipped because another run was already active. |
| `diun_watch_last_run_timestamp_seconds` | Gauge | | Unix timestamp of the last completed watch run. |
| `diun_watch_last_run_duration_seconds` | Gauge | | Duration in seconds of the last completed watch run. |
| `diun_watch_last_run_images` | Gauge | `status` | Number of images by status in the last completed watch run. |
| `diun_image_update_available` | Gauge | `provider`, `image` | `1` when the last check found an actionable update for the image, otherwise `0`. First-run baseline `new` results are not treated as actionable updates. |
| `diun_image_last_check_timestamp_seconds` | Gauge | `provider`, `image` | Unix timestamp of the last completed check for the image. |
| `diun_image_last_check_status` | Gauge | `provider`, `image`, `status` | Last check status for the image. The active status has value `1`. |
| `diun_image_created_timestamp_seconds` | Gauge | `provider`, `image` | Unix timestamp of the image manifest creation time reported by the registry. This metric is omitted if the registry does not provide a creation timestamp. |

The per-image metrics intentionally use only the `provider`, `image`, and
`status` labels to keep cardinality predictable across Docker, Swarm,
Kubernetes, Nomad, Dockerfile, and file providers.

See [Docker Compose with Prometheus metrics](../faq.md#docker-compose-with-prometheus-metrics)
for a complete Compose example.

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
