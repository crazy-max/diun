# Defaults configuration

## Overview

Defaults allow specifying default values for any configuration that is
typically set at the image level using labels or annotations depending on the
provider. Any of them will take precedence or be merged over defaults.

```yaml
defaults:
  watchRepo: false
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

### `watchRepo`

Watch all tags of this container image ([be careful](../faq.md#docker-hub-rate-limits)
with this setting). (default `false`)

!!! example "Config file"
    ```yaml
    defaults:
      watchRepo: false
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_WATCHREPO`

### `notifyOn`

List of status to be notified. Can be one of `new` or `update`.
(default `new,update`)

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
    Only works if watch repo is enabled.

!!! example "Config file"
    ```yaml
    defaults:
      maxTags: 10
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_MAXTAGS=10`

### `sortTags`

[Sort tags method](../faq.md#tags-sorting-when-using-watch_repo). Can be one of
`default`, `reverse`, `semver`, `lexicographical`. (default `reverse`)

!!! warning
    Only works if watch repo is enabled.

!!! example "Config file"
    ```yaml
    defaults:
      sortTags: reverse
    ```

!!! abstract "Environment variables"
    * `DIUN_DEFAULTS_SORTTAGS=reverse`

### `includeTags`

List of regular expressions to include tags. Can be useful if watch repo is
enabled.

!!! example "Config file"
    ```yaml
    defaults:
      includeTags:
        - ^\d+\.\d+\.\d+$
    ```

!!! abstract "Environment variables"

Comma separated list of regular expressions to include tags.

    * `DIUN_DEFAULTS_INCLUDETAGS=^\d+\.\d+\.\d+$`

### `excludeTags`

List of regular expressions to exclude tags. Can be useful if watch repo is
enabled.

!!! example "Config file"
    ```yaml
    defaults:
      excludeTags:
        - dev
    ```

!!! abstract "Environment variables"

Comma separated list of regular expressions to include tags.

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
