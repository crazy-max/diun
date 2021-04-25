# Command Line

## Usage

```shell
diun [options]
```

## Options

```
$ diun --help
Usage: diun

Docker image update notifier. More info: https://github.com/crazy-max/diun

Flags:
  -h, --help                Show context-sensitive help.
      --version
      --config=STRING       Diun configuration file ($CONFIG).
      --profiler=STRING     Profiler to use ($PROFILER).
      --log-level="info"    Set log level ($LOG_LEVEL).
      --log-json            Enable JSON logging output ($LOG_JSON).
      --log-caller          Add file:line of the caller to log output
                            ($LOG_CALLER).
      --log-nocolor         Disables the colorized output ($LOG_NOCOLOR).
      --test-notif          Test notification settings.
```

## Environment variables

Following environment variables can be used in place:

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `CONFIG`           |               | Diun configuration file |
| `PROFILER`         |               | Profiler to use |
| `LOG_LEVEL`        | `info`        | Log level output |
| `LOG_JSON`         | `false`       | Enable JSON logging output |
| `LOG_CALLER`       | `false`       | Enable to add `file:line` of the caller |
| `LOG_NOCOLOR`      | `false`       | Disables the colorized output |
