# Registry options configuration

## Overview

Registry options are used to authenticate against a registry during the
analysis of an image:

```yaml
regopts:
  - name: "myregistry"
    auth:
      static:
        username: fii
        password: bor
    timeout: 30s
  - name: "docker.io"
    selector: image
    auth:
      static:
        username: foo
        password: bar
  - name: "docker.io/crazymax"
    selector: image
    auth:
      static:
        usernameFile: /run/secrets/username
        passwordFile: /run/secrets/password
```

* `myregistry` will be used as a `name` selector (default) if referenced by its [name](#name).
* `docker.io` will be used as an `image` selector. If an image is on DockerHub (`docker.io` domain), this registry
options will be selected if not referenced as a `regopt` name.
* `docker.io/crazymax` will be used as an `image` selector. If an image is on DockerHub and in `crazymax` namespace,
this registry options will be selected if not referenced as a `regopt` name.

## Configuration

### `name`

Unique name for registry options. This name can be used through `diun.regopt`
[Docker](../providers/docker.md#docker-labels) / [Containerd](../providers/containerd.md#containerd-labels) / [Swarm](../providers/swarm.md#docker-labels)
label or [Kubernetes](../providers/kubernetes.md#kubernetes-annotations) / [Nomad](../providers/nomad.md#nomad-annotations)
annotation and also as `regopt` for the [Dockerfile](../providers/dockerfile.md)
and [File](../providers/file.md) providers.

!!! warning
    * **Required**
    * Must be **unique**

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_NAME`

### `selector`

What kind of selector to use to retrieve registry options. (default `name`)

!!! warning
    * Accepted values are `name` or `image`

* `name` selector is the default value and will retrieve this registry options only if it's referenced by its [name](#name).
* `image` selector will retrieve this registry options if the given image matches the registry domain or repository path.

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        selector: name
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_SELECTOR`

### `auth.static`

Use fixed registry credentials, supplied directly or read from secret files.
`username` takes precedence over `usernameFile`, and `password` takes precedence
over `passwordFile`. A username or username file is required. File contents
are read when creating jobs for each scan.

```yaml
regopts:
  - name: myregistry
    auth:
      static:
        username: foo
        passwordFile: /run/secrets/registry-password
```

Environment variables:

* `DIUN_REGOPTS_<KEY>_AUTH_STATIC_USERNAME`
* `DIUN_REGOPTS_<KEY>_AUTH_STATIC_USERNAMEFILE`
* `DIUN_REGOPTS_<KEY>_AUTH_STATIC_PASSWORD`
* `DIUN_REGOPTS_<KEY>_AUTH_STATIC_PASSWORDFILE`

When no authentication provider is configured, Diun discovers credentials from
Docker configuration files and credential helpers.

### `timeout`

Timeout is the maximum amount of time for the TCP connection to establish. (default `0` ; no timeout)

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        timeout: 30s
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_TIMEOUT`

### `insecureTLS`

Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification. (default `false`)

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        insecureTLS: false
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_INSECURETLS`
