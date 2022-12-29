# Nomad provider

## About

The Nomad provider allows you watch tasks running using the Docker provider for registry updates.

## Quick start

Here we'll go over basic deployment using a local Nomad cluster.

First, we'll deploy a Diun job:

```hcl
job "diun" {
  type = "service"

  group "diun" {
    task "diun" {
      driver = "docker"

      config {
        image = "crazymax/diun:latest"
        args = ["serve"]
      }

      env = {
        "NOMAD_ADDR" = "http://${attr.unique.network.ip-address}:4646/",
        "DIUN_PROVIDERS_NOMAD" = true,
      }
    }
  }
}
```

Task based configuration can be passed through Service tags or meta attributes. These can be defined at the task or even the group level where it will apply to all tasks within the group.

The example below will show all methods, but you only need to use one.

```hcl
job "whoami" {
  type = "service"

  group "whoami" {
    network {
      mode = "bridge"

      port "web" {
        to = 80
      }
    }

    // This
    meta {
      diun.enable = true
    }

    // Or this
    service {
      tags = [
        "diun.enable=true"
      ]
    }

    task "diun" {
      driver = "docker"

      config {
        image = "containous/whoami:latest"
      }

      // Or this
      meta {
        diun.enable = true
      }

      // Or this
      service {
        tags = [
          "diun.enable=true"
        ]
      }
    }
  }
}
```

## Configuration

!!! hint
    Environment variable `DIUN_PROVIDERS_NOMAD=true` can be used to enable this provider with default values.

Default values are assigned by the Nomad client. If not provided in your Diun configuration, the client will default to using the same config values as the `nomad` cli client.

!!! abstract "Environment variables"
    * `NOMAD_ADDR`
    * `NOMAD_REGION`
    * `NOMAD_NAMESPACE`
    * `NOMAD_HTTP_AUTH`
    * `NOMAD_CACERT`
    * `NOMAD_CAPATH`
    * `NOMAD_CLIENT_CERT`
    * `NOMAD_CLIENT_KEY`
    * `NOMAD_TLS_SERVER_NAME`
    * `NOMAD_SKIP_VERIFY`
    * `NOMAD_TOKEN`


### `address`

The Nomad server address as URL.

!!! example "File"
    ```yaml
    providers:
      nomad:
        address: "http://localhost:4646"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_ENDPOINT`

Nomad server endpoint as URL, which is only used when the behavior based on environment variables described below
does not apply.

### `region`

Nomad region to query from

!!! example "File"
    ```yaml
    providers:
      nomad:
        region: "region1"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_REGION`

### `namespace`

Nomad namespace to query from

!!! example "File"
    ```yaml
    providers:
      nomad:
        namespace: "namespace1"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_NAMESPACE`

### `secretID`

SecretID to connect to Nomad API. This token must have permission to query and view Nomad jobs.

!!! example "File"
    ```yaml
    providers:
      nomad:
        secretID: "secret"
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_SECRETID`

### `tlsInsecure`

Controls whether client does not verify the server's certificate chain and hostname (default `false`).

!!! example "File"
    ```yaml
    providers:
      nomad:
        tlsInsecure: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_TLSINSECURE`

### `watchByDefault`

Enable watch by default. If false, tasks that don't have `diun.enable = true` in their meta or service tags will be ignored
(default `false`).

!!! example "File"
    ```yaml
    providers:
      nomad:
        watchByDefault: false
    ```

!!! abstract "Environment variables"
    * `DIUN_PROVIDERS_NOMAD_WATCHBYDEFAULT`

## Nomad annotations

You can configure more finely the way to analyze the image of your tasks through Nomad meta attributes or service tags:

| Name                | Default      | Description                                                                                                                                             |
|---------------------|--------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `diun.enable`       |              | Set to true to enable image analysis of this task                                                                                                       |
| `diun.regopt`       |              | [Registry options](../config/regopts.md) name to use                                                                                                    |
| `diun.watch_repo`   | `false`      | Watch all tags of this task image ([be careful](../faq.md#docker-hub-rate-limits) with this setting)                                                    |
| `diun.notify_on`    | `new;update` | Semicolon separated list of status to be notified: `new`, `update`.                                                                                     |
| `diun.sort_tags`    | `reverse`    | [Sort tags method](../faq.md#tags-sorting-when-using-watch_repo) if `diun.watch_repo` enabled. One of `default`, `reverse`, `semver`, `lexicographical` |
| `diun.max_tags`     | `0`          | Maximum number of tags to watch if `diun.watch_repo` enabled. `0` means all of them                                                                     |
| `diun.include_tags` |              | Semicolon separated list of regular expressions to include tags. Can be useful if you enable `diun.watch_repo`                                          |
| `diun.exclude_tags` |              | Semicolon separated list of regular expressions to exclude tags. Can be useful if you enable `diun.watch_repo`                                          |
| `diun.hub_link`     | _automatic_  | Set registry hub link for this image                                                                                                                    |
| `diun.platform`     | _automatic_  | Platform to use (e.g. `linux/amd64`)                                                                                                                    |
