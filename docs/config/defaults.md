# Defaults configuration

## Overview

Defaults allow specifying default values for any configuration that is
typically set at the image level using labels or annotations depending on the
provider. Image-level values usually replace defaults when set. Metadata values
are merged with defaults, and provider-specific documentation calls out any
other fields that merge instead of replacing defaults.

```yaml
defaults:
  watchMode: tag
  notifyOn:
    - new
    - update
  maxTags: 10
  sortTags: reverse
  includeTags:
    - latest
  excludeTags:
    - dev
  metadata:
    foo: bar
```

## Configuration

### `watchMode`

Select which image tags to watch. (default `tag`)

| Value                | Behavior                                        |
|----------------------|-------------------------------------------------|
| `tag`                | Watch only the configured tag or digest         |
| `all`                | Watch all repository tags                       |
| `semver`             | Watch newer stable semantic version tags        |
| `semver+prereleases` | Watch newer version tags, including prereleases |

Every mode watches the configured image. Repository tags are sorted and filtered
before `maxTags` is applied. Manifest checks consume
[registry quota](../faq.md#docker-hub-rate-limits).

Semver modes require a full version (`1.5.0` or `v1.5.0`) and select strictly
newer tags. Docker variant suffixes such as `-alpine` are interpreted as
prereleases; use `semver+prereleases` and tag filters for these images. An invalid
or missing version logs an error and skips repository tags.

Per-image watch modes override defaults.

!!! example "Config file"
    ```yaml
    defaults:
      watchMode: tag
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_WATCHMODE=tag`

### `notifyOn`

List of image statuses that trigger notifications. (default `new,update`)

* `new`: The image is not yet recorded in Diun's database.
* `update`: The image on the registry has changed since the previous scan.
* `unchange`: The image on the registry has not changed since the previous scan.

!!! example "Config file"
    ```yaml
    defaults:
      notifyOn:
        - new
        - update
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_NOTIFYON=new,update`

### `maxTags`

Maximum number of tags to watch. `0` means all of them. (default `0`)

!!! warning
    Only applies when watching repository tags.

!!! example "Config file"
    ```yaml
    defaults:
      maxTags: 10
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_MAXTAGS=10`

### `sortTags`

[Sort tags method](../faq.md#tags-sorting-when-using-watch_mode). Can be one of
`default`, `reverse`, `semver`, `lexicographical`. (default `reverse`)

!!! warning
    Only applies when watching repository tags.

!!! example "Config file"
    ```yaml
    defaults:
      sortTags: reverse
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_SORTTAGS=reverse`

### `includeTags`

List of regular expressions to include tags. Can be useful when watching
repository tags.

!!! example "Config file"
    ```yaml
    defaults:
      includeTags:
        - ^\d+\.\d+\.\d+$
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_INCLUDETAGS=^\d+\.\d+\.\d+$`

### `excludeTags`

List of regular expressions to exclude tags. Can be useful when watching
repository tags.

The [file provider](../providers/file.md) merges image-level `exclude_tags`
with these defaults. Default patterns come first, followed by image patterns,
with duplicates removed. Image-level `include_tags` still replace defaults.

!!! example "Config file"
    ```yaml
    defaults:
      excludeTags:
        - dev
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_EXCLUDETAGS=dev`

### `metadata`

Additional metadata that can be used in [notification template](../faq.md#notification-template)

!!! example "Config file"
    ```yaml
    defaults:
      metadata:
        foo: bar
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_METADATA_FOO=bar`
