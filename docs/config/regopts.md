# Registry options configuration

## Overview

Registry options is used to authenticate against a registry during the analysis of an image:

```yaml
regopts:
  - name: "myregistry"
    username: fii
    password: bor
    timeout: 30s
  - name: "docker.io"
    selector: image
    username: foo
    password: bar
  - name: "docker.io/crazymax"
    selector: image
    usernameFile: /run/secrets/username
    passwordFile: /run/secrets/password
```

`myregistry` will be used as a `name` selector (default) if referenced by its [name](#name).

`docker.io` will be used as an `image` selector. If an image is on DockerHub (`docker.io` domain), this registry options will
be selected if not referenced as a `regopt` name.

`docker.io/crazymax` will be used as an `image` selector. If an image is on DockerHub and in `crazymax` namespace, this registry options will
be selected if not referenced as a `regopt` name.

## Configuration

### `name`

Unique name for registry options. This name can be used through `diun.regopt`
[Docker](../providers/docker.md#docker-labels) / [Swarm](../providers/swarm.md#docker-labels) label
or [Kubernetes annotation](../providers/kubernetes.md#kubernetes-annotations) and also as `regopt` for the
[Dockerfile](../providers/dockerfile.md) and [File](../providers/file.md) providers.

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

### `username`

Registry username.

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        username: foo
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_USERNAME`

### `usernameFile`

Use content of secret file as registry username if `username` not defined.

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        usernameFile: /run/secrets/username
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_USERNAMEFILE`

### `password`

Registry password.

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        username: foo
        password: bar
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_PASSWORD`

### `passwordFile`

Use content of secret file as registry password if `password` not defined.

!!! example "Config file"
    ```yaml
    regopts:
      - name: "myregistry"
        passwordFile: /run/secrets/password
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<KEY>_PASSWORDFILE`

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
