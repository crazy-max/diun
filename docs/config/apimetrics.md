# API Metrics configuration

## Overview

```yaml
apimetrics:
  enableApi: true
  enableScan: true
  token: ApiToken
  port: 6080
  apiPath: /v1/metrics
  scanPath: /v1/scan
```

## Configuration

### `token`

Authentication Bearer Token used for accessing the Metrics API endpoint and the Scan API endpoint. (default: ApiToken)

!!! example "Config file"
    ```yaml
    apimetrics:
      token: ApiToken
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_TOKEN`

### `port`

TCP port used for the http api.  (default: 6080)

!!! example "Config file"
    ```yaml
    apimetrics:
      port: ApiToken
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_PORT`

### `enableApi`

Enable or disable the Metrics endpoint. (default: false)

!!! example "Config file"
    ```yaml
    apimetrics:
      enableApi: true
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_ENABLEAPI`

### `apiPath`

Path to expose the API Metrics on. (default: /v1/metrics)

!!! example "Config file"
    ```yaml
    apimetrics:
      apiPath: /v1/metrics
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_APIPATH`

### `enableScan`

Enable or disable the Scan endpoint.  The Scan endpoint allows for the triggering of a re-scan of the images. (default: false)

!!! example "Config file"
    ```yaml
    apimetrics:
      enableScan: true
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_APISCAN`

### `scanPath`

Path to expose the API Metrics on. (default: /v1/scann)

!!! example "Config file"
    ```yaml
    apimetrics:
      apiScan: true
    ```

!!! abstract "Environment variables"
    * `DIUN_APIMETRICS_APISCAN`
