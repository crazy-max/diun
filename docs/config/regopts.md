# Registries options configuration

## `username`

Registry username.

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        username: foo
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_USERNAME`

## `usernameFile`

Use content of secret file as registry username if `username` not defined.

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        usernameFile: /run/secrets/username
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_USERNAMEFILE`

## `password`

Registry password.

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        username: foo
        password: bar
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_PASSWORD`

## `passwordFile`

Use content of secret file as registry password if `password` not defined.

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        usernameFile: /run/secrets/username
        usernameFile: /run/secrets/password
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_PASSWORDFILE`

## `timeout`

Timeout is the maximum amount of time for the TCP connection to establish. (default `10s`)

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_TIMEOUT`

## `insecureTLS`

Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification. (default `false`)

!!! example "Config file"
    ```yaml
    regopts:
      <name>:
        insecureTLS: false
    ```

!!! abstract "Environment variables"
    * `DIUN_REGOPTS_<NAME>_INSECURETLS`
