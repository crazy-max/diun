# Upgrade notes

## 2.x > 3.x



## 1.x > 2.x

`image` field has been moved to `providers.static` in configuration file:

> **1.x**
```yaml
image:
  - name: docker.io/crazymax/diun
    watch_repo: true
    max_tags: 10
```

> **2.x**
```yaml
providers:
  static:
    - name: docker.io/crazymax/diun
      watch_repo: true
      max_tags: 10
```

See [providers configuration](doc/configuration.md#providers) for more info.

## 0.x > 1.x

Some fields in configuration file has been changed:

* `registries` renamed `regopts`
* `items` renamed `image`
* `items[].image` renamed `image[].name`
* `items[].registry_id` renamed `image[].regopts_id`
* `watch.os` and `watch.arch` moved to `image[].os` and `image[].arch`

> **0.x**
```yaml
watch:
  os: linux
  arch: amd64

registries:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20

items:
  - image: docker.io/crazymax/nextcloud:latest
    registry_id: someregistryoptions
```

> **1.x**
```yaml
regopts:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20

image:
  - name: docker.io/crazymax/nextcloud:latest
    regopts_id: someregistryoptions
    os: linux
    arch: amd64
```
