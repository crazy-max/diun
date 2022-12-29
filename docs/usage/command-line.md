# Command Line

## Usage

```shell
diun [global options] command [command or global options] [arguments...]
```

## Global options

All global options can be placed at the command level.

* `--help`, `-h`: Show context-sensitive help.
* `--version`: Show version and exit.

## Commands

### `serve`

Starts Diun server.

* `--config <path>`: Diun configuration file
* `--profiler-path <path>`: Base path where profiling files are written
* `--profiler <string>`: Profiler to use
* `--log-level <string>`: Set log level (default `info`)
* `--log-json`: Enable JSON logging output
* `--log-caller`: Add `file:line` of the caller to log output
* `--log-nocolor`: Disables the colorized output
* `--grpc-authority <string>`: Address used to expose the gRPC server (default `:42286`)

Examples:

```shell
diun serve --config diun.yml --log-level debug
```

Following environment variables can also be used in place:

| Name             | Default  | Description                                                          |
|------------------|----------|----------------------------------------------------------------------|
| `CONFIG`         |          | Diun configuration file                                              |
| `PROFILER_PATH`  |          | Base path where profiling files are written                          |
| `PROFILER`       |          | [Profiler](../faq.md#profiling) to use                               |
| `LOG_LEVEL`      | `info`   | One of `panic`, `fatal`, `error`, `warn`, `info`, `debug` or `trace` |
| `LOG_JSON`       | `false`  | Enable JSON logging output                                           |
| `LOG_CALLER`     | `false`  | Enable to add `file:line` of the caller                              |
| `LOG_NOCOLOR`    | `false`  | Disables the colorized output                                        |
| `GRPC_AUTHORITY` | `:42286` | Address used to expose the gRPC server                               |

### `image list`

!!! note
    Diun needs to be started through [`serve`](#serve) command to be able to use this command.

List images in database.

* `--raw`: JSON output
* `--grpc-authority <string>`: Link to Diun gRPC API (default `127.0.0.1:42286`)

Examples:

```shell
diun image list
```
```shell
diun image list --raw
```

### `image inspect`

!!! note
    Diun needs to be started through [`serve`](#serve) command to be able to use this command.

Display information of an image in database.

* `--image`: Image to inspect (**required**)
* `--raw`: JSON output
* `--grpc-authority <string>`: Link to Diun gRPC API (default `127.0.0.1:42286`)

Examples:

```shell
diun image inspect --image alpine
```
```shell
diun image inspect --image drone/drone --raw
```

### `image remove`

!!! note
    Diun needs to be started through [`serve`](#serve) command to be able to use this command.

Remove an image manifest from database.

* `--image`: Image to remove (**required**)
* `--grpc-authority <string>`: Link to Diun gRPC API (default `127.0.0.1:42286`)

Examples:

```shell
diun image remove --image alpine:latest
```
```shell
diun image inspect --image drone/drone
```

!!! warning
    All manifest for an image will be removed if no tag is specified

### `image prune`

!!! note
    Diun needs to be started through [`serve`](#serve) command to be able to use this command.

Remove all manifests from the database.

* `--force`: Do not prompt for confirmation
* `--grpc-authority <string>`: Link to Diun gRPC API (default `127.0.0.1:42286`)

Examples:

```shell
diun image prune
```

### `notif test`

!!! note
    Diun needs to be started through [`serve`](#serve) command to be able to use this command.

Test notification settings.

* `--grpc-authority <string>`: Link to Diun gRPC API (default `127.0.0.1:42286`)

Examples:

```shell
diun notif test
```
