# Upgrade notes

* [2.x > 3.x](#2x--3x)
  * [File provider](#file-provider)
  * [Allow only one Docker and Swarm provider](#allow-only-one-docker-and-swarm-provider)
  * [Remove `enable` setting for notifiers](#remove-enable-setting-for-notifiers)
* [1.x > 2.x](#1x--2x)
* [0.x > 1.x](#0x--1x)

## 2.x > 3.x

### File provider

`static` provider has been renamed `file`. This now allows the static configuration to be declared in one or more files to avoid overloading the current configuration file and also dynamic updating.

> **2.x**
```yaml
providers:
  static:
    - name: docker.io/crazymax/diun
      watch_repo: true
      max_tags: 10
```

> **3.x**
```yaml
providers:
  file:
    # Watch images from filename /path/to/config.yml
    filename: /path/to/config.yml
    # OR watch images from directory /path/to/config/folder
    directory: /path/to/config/folder
```
```yaml
# /path/to/config.yml
- name: docker.io/crazymax/diun
  watch_repo: true
  max_tags: 10
```

### Allow only one Docker and Swarm provider

Now you can declare only one Docker and/or Swarm provider.

> **2.x**
```yaml
providers:
  docker:
    mydocker:
      watch_stopped: true
  providers:
    swarm:
      myswarm:
        watch_by_default: true
```

> **3.x**
```yaml
providers:
  docker:
    watch_stopped: true
  swarm:
    watch_by_default: true
```

### Remove `enable` setting for notifiers

The `enable` entry has been removed for notifiers. If you don't want a notifier to be enabled, you must now remove its configuration.

> **2.x**
```yaml
notif:
  amqp:
    enable: false
    host: localhost
    port: 5672
  gotify:
    enable: true
    endpoint: http://gotify.foo.com
    token: Token123456
    priority: 1
    timeout: 10
```

> **3.x**
```yaml
notif:
  gotify:
    endpoint: http://gotify.foo.com
    token: Token123456
    priority: 1
    timeout: 10
```

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
